package cdp

import (
	"errors"
	"log"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// Params parameters of CDP method
type Params = map[string]interface{}

// MessageResult message result
type MessageResult = map[string]interface{}

// MessageReq CDP webSocket request
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

// CreateCDPClient create new client to interact with browser by CDP
func CreateCDPClient(webSocketURL string) (*Client, error) {
	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL, nil)
	if err != nil {
		return nil, err
	}
	client := &Client{
		conn:     conn,
		send:     make(chan MessageReq),              /* channel for message sending */
		recv:     make(map[int64]chan MessageResult), /* channel for receive message response */
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

// NewSession open new session on page target targetID
// if targetID == nil then client will find already opened targets with page type
// if no one target exists - client will create a new one
func (client *Client) NewSession(targetID *string) (*Session, error) {
	session := newSession(client)

	targets, err := session.getTargets()
	if err != nil {
		return nil, err
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
			return nil, err
		}
		if tID == "" {
			return nil, errors.New("no one target with page type")
		}
		targetID = &tID
	}

	if session.sessionID, err = session.attachToTarget(*targetID); err != nil {
		return nil, err
	}
	session.targetID = *targetID
	session.frameID = session.targetID
	client.sessions[session.sessionID] = session

	if _, err = session.blockingSend("Page.enable", &Params{}); err != nil {
		return nil, err
	}
	if _, err = session.blockingSend("Runtime.enable", &Params{}); err != nil {
		return nil, err
	}
	if err = session.NetworkEnable(); err != nil {
		return nil, err
	}
	if err = session.setLifecycleEventsEnabled(true); err != nil {
		return nil, err
	}
	return session, session.switchContext(*targetID)
}

func (client *Client) sendMethod(sessionID string, method string, params *Params) chan MessageResult {
	client.mutex.Lock()
	defer client.mutex.Unlock()
	client.messageID++
	client.recv[client.messageID] = make(chan MessageResult, 1)
	client.send <- MessageReq{ID: client.messageID, SessionID: sessionID, Method: method, Params: *params}
	return client.recv[client.messageID]
}

// Close close session
func (client *Client) Close() {
	client.sendMethod("", "Browser.close", &Params{})
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
			if strings.Contains(err.Error(), "use of closed network connection") {
				return
			}
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
