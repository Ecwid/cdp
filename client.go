package cdp

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Params параметры CDP метода
type Params = map[string]interface{}

// MessageResult результат вызова метода
type MessageResult = map[string]interface{}

// MessageReq структура CDP запроса
type MessageReq struct {
	ID        int64  `json:"id"`
	Method    string `json:"method"`
	Params    Params `json:"params"`
	SessionID string `json:"sessionId,omitempty"`
}

// ProtocolDebug chrome devtools protocol messages debug
var ProtocolDebug = false

// Client ...
type Client struct {
	mutex     sync.Mutex
	messageID int64
	conn      *websocket.Conn
	send      chan MessageReq
	recv      map[int64]chan MessageResult
	events    chan MessageResult
	close     chan bool
	sessions  map[string]*Session
}

// CreateCDPClient открывает новую сессию с браузером
func CreateCDPClient(webSocketURL string) (*Client, error) {
	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL, nil)
	if err != nil {
		return nil, err
	}
	client := &Client{
		conn:     conn,
		send:     make(chan MessageReq),              /* Канал для отправки сообщений по протоколу */
		recv:     make(map[int64]chan MessageResult), /* Канал для для получения ответа на сообщение */
		sessions: make(map[string]*Session),
		events:   make(chan MessageResult, 1),
		close:    make(chan bool),
	}
	go client.sender()
	go client.broadcast()
	client.sendMethod("", "Target.setDiscoverTargets", &Params{"discover": true})
	return client, nil
}

func (client *Client) deleteSession(sessionID string) {
	client.mutex.Lock()
	delete(client.sessions, sessionID)
	client.mutex.Unlock()
}

// NewSession ...
func (client *Client) NewSession(targetID *string) *Session {
	session := newSession(client)

	targets, err := session.getTargets()
	if err != nil {
		panic(err)
	}

	// try to find exist page target
	if targetID == nil {
		for _, t := range targets {
			if t.Type == "page" {
				targetID = &t.TargetID
				break
			}
		}
	}

	// create a new if no one page target
	if targetID == nil {
		tID, err := session.createTarget("about:blank")
		if err != nil {
			panic(err)
		}
		if tID == "" {
			panic(`No one target with page type'`)
		}
		targetID = &tID
	}

	session.sessionID, err = session.attachToTarget(*targetID)
	if err != nil {
		panic(err)
	}
	session.targetID = *targetID
	session.frameID = session.targetID
	client.sessions[session.sessionID] = session

	_, _ = session.blockingSend("Page.enable", &Params{})
	_, _ = session.blockingSend("Runtime.enable", &Params{})
	session.NetworkEnable()
	_ = session.setLifecycleEventsEnabled(true)

	session.switchContext(*targetID)
	return session
}

func (client *Client) sendMethod(sessionID string, method string, params *Params) chan MessageResult {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	client.messageID++
	client.recv[client.messageID] = make(chan MessageResult, 1)
	client.send <- MessageReq{ID: client.messageID, SessionID: sessionID, Method: method, Params: *params}
	return client.recv[client.messageID]
}

// Close ...
func (client *Client) Close() {
	close(client.close)
}

func (client *Client) sender() {
	defer client.conn.Close()
	for {
		select {
		case req := <-client.send:
			if ProtocolDebug {
				log.Printf("\033[1;36msend -> %+v\033[0m", req)
			}
			err := client.conn.WriteJSON(req)
			if err != nil {
				log.Printf(err.Error())
				break
			}
		case <-client.close:
			_ = client.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			close(client.events)
			return
		}
	}
}

func (client *Client) broadcast() {
	var message MessageResult
	var err error
	for {
		message = make(MessageResult)
		err = client.conn.ReadJSON(&message)
		if err != nil {
			log.Printf(err.Error())
			break
		}
		if id, has := message["id"]; has {
			if ProtocolDebug {
				if _, e := message["error"]; e {
					log.Printf("\033[1;31mrecv <- %+v\033[0m", message)
				} else {
					log.Printf("\033[1;32mrecv <- %+v\033[0m", message)
				}
			}
			messageID := int64(id.(float64))
			if recv, ok := client.recv[messageID]; ok {
				recv <- message
				client.mutex.Lock()
				delete(client.recv, messageID)
				client.mutex.Unlock()
			}
		} else {
			if ProtocolDebug {
				log.Printf("recv <- %+v", message)
			}
			if session, has := message["sessionId"].(string); has {
				client.sessions[session].incomingEvent <- message
			} else {
				// Если у входящего сообщения нет sessionID, то отправим его всем сессиям
				for _, e := range client.sessions {
					e.incomingEvent <- message
				}
			}
		}
		message = nil
	}
}
