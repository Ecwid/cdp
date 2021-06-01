package cdp

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/ecwid/cdp/pkg/devtool"
)

// Map ...
type Map map[string]interface{}

// Session ...
type Session struct {
	ws              *WSClient
	id              string
	state           *state
	target          string
	broadcast       chan *wsBroadcast
	closed          chan struct{}
	err             chan error
	deadline        time.Duration
	eventsMutex     *sync.Mutex
	listeners       map[string]*list.List
	OnElementBinded func(e *Element)
}

func newSession(ws *WSClient) *Session {
	return &Session{
		id:              "",
		ws:              ws,
		eventsMutex:     &sync.Mutex{},
		state:           newState(),
		listeners:       map[string]*list.List{},
		broadcast:       make(chan *wsBroadcast, 10),
		closed:          make(chan struct{}, 1),
		err:             make(chan error, 1),
		deadline:        60 * time.Second,
		OnElementBinded: nil,
	}
}

// NewSession ...
func NewSession(session *Session, target string) (*Session, error) {
	newsess := newSession(session.ws)
	err := newsess.attachToTarget(target)
	return newsess, err
}

// ID session's ID
func (session Session) ID() string {
	return session.target
}

func (session Session) exception(err error) {
	session.ws.printf(LevelProtocolFatal, err.Error())
	select {
	case <-session.closed:
		return
	case session.err <- err:
	default:
	}
}

func (session *Session) attachToTarget(targetID string) error {
	var result = make(Map)
	err := session.call("Target.attachToTarget", Map{"targetId": targetID, "flatten": true}, &result)
	if err != nil {
		return err
	}
	session.target = targetID
	session.id = result["sessionId"].(string)
	session.ws.register(session.id, session.broadcast)
	go session.listener()

	session.state.reset()

	if err := session.call("Target.setDiscoverTargets", Map{"discover": true}, nil); err != nil {
		return err
	}
	if err = session.call("Page.enable", nil, nil); err != nil {
		return err
	}
	if err = session.call("Runtime.enable", nil, nil); err != nil {
		return err
	}
	// maxPostDataSize - Longest post body size (in bytes) that would be included in requestWillBeSent notification
	if err = session.call("Network.enable", Map{"maxPostDataSize": 2 * 1024}, nil); err != nil {
		return err
	}
	return nil
}

func (session Session) listener() {
	defer func() {
		close(session.closed)
		session.ws.unregister(session.id)
	}()
	for e := range session.broadcast {

		if e.Error != "" {
			session.exception(errors.New(e.Error))
			return
		}

		session.eventsMutex.Lock()
		if list, has := session.listeners[e.Method]; has {
			for p := list.Front(); p != nil; p = p.Next() {
				p.Value.(func(*Event))(&e.Event)
			}
		}
		session.eventsMutex.Unlock()

		switch e.Method {

		case "Runtime.executionContextCreated":
			c := new(devtool.ExecutionContextCreated)
			if err := json.Unmarshal(e.Params, c); err != nil {
				session.exception(err)
			}
			var fid = c.Context.AuxData["frameId"].(string)
			if fid == session.state.GetFrame() {
				session.state.set(fid, c.Context.ID)
			}

		case "Target.targetCrashed":
			session.exception(errors.New(string(e.Params)))
			// ws client will be disconnected
			return

		case "Target.targetDestroyed":
			event := new(devtool.TargetDestroyed)
			if err := json.Unmarshal(e.Params, event); err != nil {
				session.exception(err)
				return
			}
			if event.TargetID == session.target {
				// stop listener
				return
			}

		case "Target.detachedFromTarget":
			event := new(devtool.DetachedFromTarget)
			if err := json.Unmarshal(e.Params, event); err != nil {
				session.exception(err)
				return
			}
			if event.SessionID == session.id {
				// stop listener
				return
			}
		}
	}
}

func (session Session) blockingSend(method string, params interface{}) ([]byte, error) {
	recv := session.ws.sendOverProtocol(session.id, method, params)
	select {
	case err := <-session.err:
		return nil, err
	case <-session.closed:
		return nil, ErrSessionAlreadyClosed
	case response := <-recv:
		if response.Error.Code != 0 {
			switch response.Error.Message {
			case "Cannot find context with specified id":
				return nil, ErrStaleElementReference
			default:
				return nil, response.Error
			}
		}
		return response.Result, nil
	case <-time.After(session.deadline):
		return nil, fmt.Errorf("websocket response timeout was reached %s for %s(%+v)", session.deadline.String(), method, params)
	}
}

func (session Session) call(method string, req interface{}, resp interface{}) error {
	b, err := session.blockingSend(method, req)
	if err != nil {
		return err
	}
	if resp == nil || (reflect.ValueOf(resp).Kind() == reflect.Ptr && reflect.ValueOf(resp).IsNil()) {
		return nil
	}
	return json.Unmarshal(b, resp)
}

// Subscribe subscribe to CDP event
func (session *Session) Subscribe(method string, cb func(event *Event)) (unsubscribe func()) {
	session.eventsMutex.Lock()
	defer session.eventsMutex.Unlock()
	if _, has := session.listeners[method]; !has {
		session.listeners[method] = list.New()
	}
	p := session.listeners[method].PushBack(cb)
	return func() {
		session.eventsMutex.Lock()
		defer session.eventsMutex.Unlock()
		session.listeners[method].Remove(p)
	}
}

// GetTargets ....
func (session Session) GetTargets() ([]*devtool.TargetInfo, error) {
	var info = new(devtool.TargetInfos)
	if err := session.call("Target.getTargets", nil, &info); err != nil {
		return nil, err
	}
	return info.TargetInfos, nil
}

// GetID ...
func (session Session) GetID() string {
	return session.id
}

// SetTimeout ...
func (session *Session) SetTimeout(dl time.Duration) {
	session.deadline = dl
}

// GetTimeout ...
func (session Session) GetTimeout() time.Duration {
	return session.deadline
}

// SetOutLevel ...
func (session *Session) SetOutLevel(level OutLevel) {
	session.ws.outLevel = level
}

// Close close this sessions
func (session Session) Close() error {
	err := session.call("Target.closeTarget", Map{"targetId": session.target}, nil)
	// event 'Target.targetDestroyed' can be received early than message response
	if err == ErrSessionAlreadyClosed {
		return nil
	}
	return err
}

// IsClosed check is session (tab) closed
func (session Session) IsClosed() bool {
	select {
	case <-session.ws.disconnected:
		return true
	case <-session.closed:
		return true
	default:
		return false
	}
}
