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
	ws        *WSClient
	id        string
	targetID  string
	frameID   string
	rw        *sync.Mutex
	frames    *sync.Map
	listeners map[string]*list.List
	broadcast chan *wsBroadcast
	closed    chan struct{}
	err       chan error
	deadline  time.Duration
}

func newSession(ws *WSClient) *Session {
	return &Session{
		id:        "",
		ws:        ws,
		rw:        &sync.Mutex{},
		frames:    &sync.Map{},
		listeners: map[string]*list.List{},
		broadcast: make(chan *wsBroadcast, 10),
		closed:    make(chan struct{}, 1),
		err:       make(chan error, 1),
		deadline:  60 * time.Second,
	}
}

// Session ...
func (c Browser) Session() (*Session, error) {
	var sess = newSession(c.GetWSClient())
	targets, err := sess.GetTargets()
	if err != nil {
		return nil, err
	}
	for _, t := range targets {
		if t.Type == "page" {
			return sess, sess.start(t.TargetID)
		}
	}
	return nil, ErrNoPageTarget
}

// NewSession ...
func NewSession(session *Session, target string) (*Session, error) {
	newsess := newSession(session.ws)
	err := newsess.start(target)
	return newsess, err
}

// ID session's ID
func (session Session) ID() string {
	return session.id
}

func (session Session) close(err error) {
	select {
	case <-session.closed:
		return
	case session.err <- err:
	default:
	}
	close(session.closed)
}

func (session *Session) start(targetID string) error {
	if err := session.call("Target.setDiscoverTargets", Map{"discover": true}, nil); err != nil {
		return err
	}
	var result = make(Map)
	err := session.call("Target.attachToTarget", Map{"targetId": targetID, "flatten": true}, &result)
	if err != nil {
		return err
	}
	session.targetID = targetID
	session.frameID = targetID
	session.id = result["sessionId"].(string)

	go session.listener()
	session.ws.subscribe(session.id, session.broadcast)
	if err = session.call("Page.enable", nil, nil); err != nil {
		return err
	}
	if err = session.call("Runtime.enable", nil, nil); err != nil {
		return err
	}
	// maxPostDataSize - Longest post body size (in bytes) that would be included in requestWillBeSent notification
	if err = session.call("Runtime.enable", Map{"maxPostDataSize": 1024}, nil); err != nil {
		return err
	}
	return nil
}

func (session Session) listener() {
	for e := range session.broadcast {

		if e.Error != "" {
			session.close(errors.New(e.Error))
			return
		}

		session.rw.Lock()
		if list, has := session.listeners[e.Method]; has {
			for p := list.Front(); p != nil; p = p.Next() {
				p.Value.(func(*Event))(&e.Event)
			}
		}
		session.rw.Unlock()

		switch e.Method {
		case "Runtime.executionContextsCleared":
			session.frames.Range(func(k interface{}, v interface{}) bool {
				session.frames.Delete(k)
				return true
			})

		case "Runtime.executionContextCreated":
			c := new(devtool.ExecutionContextCreated)
			if err := json.Unmarshal(e.Params, c); err != nil {
				session.close(err)
				return
			}
			session.frames.Store(c.Context.AuxData["frameId"].(string), c.Context.ID)

		case "Runtime.executionContextDestroyed":
			c := new(devtool.ExecutionContextDestroyed)
			if err := json.Unmarshal(e.Params, c); err != nil {
				session.close(err)
				return
			}
			session.frames.Range(func(k interface{}, v interface{}) bool {
				if v.(int64) == c.ExecutionContextID {
					session.frames.Delete(k)
					return false
				}
				return true
			})

		case "Target.targetCrashed":
			c := new(devtool.TargetCrashed)
			if err := json.Unmarshal(e.Params, c); err != nil {
				session.close(err)
				return
			}
			session.close(errors.New(string(e.Params)))
			return

		case "Target.targetDestroyed":
			c := new(devtool.TargetDestroyed)
			if err := json.Unmarshal(e.Params, c); err != nil {
				session.close(err)
				return
			}
			if c.TargetID == session.targetID {
				session.close(nil)
				session.ws.unsubscribe(session.id)
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
		return nil, ErrSessionClosed
	case response := <-recv:
		if response.Error.Code != 0 {
			return nil, response.Error
		}
		return response.Result, nil
	case <-time.After(session.deadline):
		return nil, fmt.Errorf("websocket response timeout was reached %s for %s(%+v)", session.deadline.String(), method, params)
	}
}

func (session Session) withDeadline(c chan struct{}) error {
	select {
	case <-c:
		return nil
	case err := <-session.err:
		return err
	case <-session.closed:
		return ErrSessionClosed
	case <-time.After(session.deadline):
		return ErrTimeout
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
	session.rw.Lock()
	defer session.rw.Unlock()
	if _, has := session.listeners[method]; !has {
		session.listeners[method] = list.New()
	}
	p := session.listeners[method].PushBack(cb)
	return func() {
		session.rw.Lock()
		defer session.rw.Unlock()
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

func (session Session) executionContext() (int64, error) {
	var id = session.frameID
	if v, ok := session.frames.Load(id); ok {
		return v.(int64), nil
	}
	if id == session.targetID {
		// for main frame we can use default context with ID = 0
		return 0, nil
	}
	return -1, ErrFrameDetached
}

// GetID ...
func (session Session) GetID() string {
	return session.id
}

// SetTimeout ...
func (session *Session) SetTimeout(dl time.Duration) {
	session.deadline = dl
}

// Close close this sessions
func (session Session) Close() error {
	err := session.call("Target.closeTarget", Map{"targetId": session.targetID}, nil)
	// event 'Target.targetDestroyed' can be received early than message response
	if err != nil && err != ErrSessionClosed {
		return err
	}
	session.close(nil)
	return nil
}

// IsClosed check is session (tab) closed
func (session Session) IsClosed() bool {
	select {
	case <-session.closed:
		return true
	default:
		return false
	}
}

// WaitElement ...
func (session Session) WaitElement(selector string) (*Element, error) {
	var ret *Element
	return ret, NewTicker(session.deadline, 500*time.Millisecond, func() (err error) {
		ret, err = session.Query(selector)
		return err
	})
}

// NewTicker ...
func NewTicker(deadline, poll time.Duration, call func() error) (err error) {
	if err = call(); err == nil {
		return nil
	}
	var ticker = time.NewTicker(poll)
	var timeout = time.NewTimer(deadline)
	defer ticker.Stop()
	defer timeout.Stop()
	for {
		select {
		case <-timeout.C:
			return err
		case <-ticker.C:
			if err = call(); err == nil {
				return nil
			}
		}
	}
}
