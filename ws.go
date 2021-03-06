package cdp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"

	"github.com/gorilla/websocket"
)

// OutLevel ...
type OutLevel = int64

// ws log levels
const (
	LevelProtocolFatal   OutLevel = 0x01
	LevelProtocolErrors  OutLevel = 0x02 | LevelProtocolFatal
	LevelProtocolMessage OutLevel = 0x04 | LevelProtocolErrors
	LevelProtocolEvents  OutLevel = 0x08 | LevelProtocolErrors
	LevelProtocolVerbose OutLevel = LevelProtocolErrors | LevelProtocolMessage | LevelProtocolEvents
	LevelSessionState    OutLevel = 0x10 | LevelProtocolErrors
	LevelVerbose         OutLevel = 0xFF
)

// WSClient ...
type WSClient struct {
	WebSocketURL  string
	conn          *websocket.Conn
	id            int64
	queueMutex    *sync.Mutex
	sessionsMutex *sync.Mutex
	send          chan []byte
	receive       map[int64]chan *wsResponse
	listeners     map[string]chan *wsBroadcast
	disconnected  chan struct{}
	err           chan error
	out           *log.Logger
	outLevel      OutLevel
}

type wsError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

func (e wsError) Error() string {
	return e.Message
}

type jsonstr []byte

// MarshalJSON returns m as the JSON encoding of m.
func (m jsonstr) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *jsonstr) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("raw: UnmarshalJSON on nil pointer")
	}
	*m = append((*m)[0:0], data...)
	return nil
}

type wsMessage struct {
	ID        int64       `json:"id"`
	Method    string      `json:"method"`
	Params    interface{} `json:"params,omitempty"`
	SessionID string      `json:"sessionId,omitempty"`
}

type wsResponse struct {
	ID        int64   `json:"id,omitempty"`
	Result    jsonstr `json:"result,omitempty"`
	SessionID string  `json:"sessionId,omitempty"`
	Error     wsError `json:"error,omitempty"`
	Method    string  `json:"method,omitempty"`
	Params    jsonstr `json:"params,omitempty"`
}

// Event ...
type Event struct {
	Method string
	Params []byte
}

type wsBroadcast struct {
	Event
	Error string
}

func (r wsResponse) isBroadcast() bool {
	return r.ID == 0 && r.Method != ""
}

func (r wsResponse) isError() bool {
	return r.Error.Code != 0
}

// NewWebSocketClient ...
func NewWebSocketClient(webSocketURL string) (*WSClient, error) {
	conn, _, err := websocket.DefaultDialer.Dial(webSocketURL, nil)
	if err != nil {
		return nil, err
	}
	ws := &WSClient{
		WebSocketURL:  webSocketURL,
		conn:          conn,
		queueMutex:    &sync.Mutex{},
		sessionsMutex: &sync.Mutex{},
		send:          make(chan []byte),
		receive:       make(map[int64]chan *wsResponse),
		disconnected:  make(chan struct{}, 1),
		listeners:     make(map[string]chan *wsBroadcast),
		id:            0,
		out:           log.New(os.Stderr, "", log.LstdFlags),
		outLevel:      LevelProtocolErrors,
	}
	go ws.writer()
	go ws.reader()
	return ws, nil
}

// SetLogOutput ...
func (w *WSClient) SetLogOutput(writer io.Writer) {
	w.out.SetOutput(writer)
}

// SetLogLevel ...
func (w *WSClient) SetLogLevel(level OutLevel) {
	w.outLevel = level
}

func (w WSClient) printf(level OutLevel, format string, v ...interface{}) {
	if level&w.outLevel == level {
		_, fn, line, _ := runtime.Caller(1)
		w.out.Printf("%s:%d %s", fn, line, fmt.Sprintf(format, v...))
	}
}

func (w *WSClient) sendOverProtocol(sessionID string, method string, params interface{}) chan *wsResponse {
	w.queueMutex.Lock()
	w.id++
	response := make(chan *wsResponse, 1)
	w.receive[w.id] = response
	w.queueMutex.Unlock()

	request, err := json.Marshal(wsMessage{
		ID:        w.id,
		SessionID: sessionID,
		Method:    method,
		Params:    params,
	})
	if err != nil {
		w.printf(LevelProtocolFatal, err.Error())
		response <- &wsResponse{Error: wsError{Message: err.Error()}}
		return response
	}

	select {
	case w.send <- request:
	case <-w.disconnected:
	}

	return response
}

// Subscribe ...
func (w *WSClient) register(sessionID string, events chan *wsBroadcast) {
	w.sessionsMutex.Lock()
	defer w.sessionsMutex.Unlock()
	w.listeners[sessionID] = events
}

// Unsubscribe ...
func (w *WSClient) unregister(sessionID string) {
	w.sessionsMutex.Lock()
	defer w.sessionsMutex.Unlock()
	delete(w.listeners, sessionID)
}

// Close ...
func (w *WSClient) exception(err error) {
	if err != nil {
		w.printf(LevelProtocolFatal, "\033[1;31m%s\033[0m", err.Error())
	}
	w.err <- err
}

func (w *WSClient) writer() {
	for {
		select {
		case <-w.disconnected:
			_ = w.conn.Close()
			w.publish(&wsResponse{Error: wsError{Message: ErrConnectionClosed.Error()}})
			return
		case req := <-w.send:
			w.printf(LevelProtocolMessage, "\033[1;36msend -> %s\033[0m", string(req))
			if err := w.conn.WriteMessage(websocket.TextMessage, req); err != nil {
				w.exception(err)
			}
		case err := <-w.err:
			if err != nil {
				w.publish(&wsResponse{Error: wsError{Message: err.Error()}})
			}
			return
		}
	}
}

func (w *WSClient) publish(response *wsResponse) {
	var event = &wsBroadcast{
		Event: Event{
			Method: response.Method,
			Params: response.Params,
		},
		Error: response.Error.Message,
	}
	w.sessionsMutex.Lock()
	defer w.sessionsMutex.Unlock()
	if response.SessionID != "" {
		w.listeners[response.SessionID] <- event
	} else {
		for _, v := range w.listeners {
			v <- event
		}
	}
}

func (w *WSClient) reader() {
	for {
		_, body, err := w.conn.ReadMessage()
		if err != nil {
			switch err.(type) {
			case *websocket.CloseError:
				close(w.disconnected)
				// do nothing, browser was closed
				return
			default:
				w.exception(err)
				return
			}
		}
		var response = new(wsResponse)
		if err := json.Unmarshal(body, response); err != nil {
			w.exception(err)
			return
		}
		if response.isBroadcast() {
			w.printf(LevelProtocolEvents, "\033[1;30mevent <- %s\033[0m", string(body))
			w.publish(response)
		} else {
			if response.isError() {
				w.printf(LevelProtocolErrors, "\033[1;31mrecv_err <- %s\033[0m", string(body))
			} else {
				w.printf(LevelProtocolMessage, "\033[1;34mrecv <- %s\033[0m", string(body))
			}
			w.queueMutex.Lock()
			if recv, has := w.receive[response.ID]; has {
				recv <- response
				delete(w.receive, response.ID)
			}
			w.queueMutex.Unlock()
		}
	}
}
