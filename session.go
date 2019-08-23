package cdp

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// NavigationTimeout - navigation timeout
var NavigationTimeout = time.Second * 40

// WebSocketTimeout - web socket response timeout
var WebSocketTimeout = time.Minute * 3

// ErrSessionClosed current session (target) already closed
var ErrSessionClosed = errors.New(`Session closed`)

// Session CDP session
type Session struct {
	rw            sync.RWMutex
	client        *Client
	sessionID     string
	targetID      string
	contextID     int64
	frameID       string
	incomingEvent chan MessageResult // queue of incoming events from browser
	callbacks     map[string][]func(Params)
	closed        chan bool
}

func newSession(client *Client) *Session {
	session := &Session{
		client:        client,
		incomingEvent: make(chan MessageResult, 2000),
		callbacks:     make(map[string][]func(Params)),
		closed:        make(chan bool, 1),
	}
	go session.listener()
	return session
}

// Client get client associated with this session
func (session *Session) Client() *Client {
	return session.client
}

func (session *Session) switchContext(frameID string) {
	session.frameID = frameID
	var err error
	session.contextID, err = session.createIsolatedWorld(frameID)
	if err != nil {
		panic(err)
	}
}

func unmarshal(source interface{}, dest interface{}) {
	str, err := json.Marshal(source)
	if err != nil {
		panic(err) // Fatal Error in Protocol
	}
	if err = json.Unmarshal([]byte(str), dest); err != nil {
		panic(err) // Fatal Error in Protocol
	}
}

// ID session's id
func (session *Session) ID() string {
	return session.sessionID
}

// SwitchToFrame switch context to frame
func (session *Session) SwitchToFrame(selector string) error {
	el, err := session.findElement(selector)
	if err != nil {
		return err
	}
	node, err := session.describeNode(el)
	if err != nil {
		return err
	}
	if "IFRAME" != node.NodeName {
		return errors.New(`Selector ` + selector + ` must be IFRAME`)
	}
	session.switchContext(node.FrameID)
	return nil
}

// MainFrame switch context to main frame
func (session *Session) MainFrame() {
	session.switchContext(session.targetID)
}

// Navigate navigate to
func (session *Session) Navigate(urlStr string) error {
	eventFired := make(chan bool)
	unsubscribe := session.Subscribe("Page.domContentEventFired", func(params Params) {
		eventFired <- true
	})
	defer unsubscribe()
	nav, err := session.navigate(urlStr, session.frameID)
	if err != nil {
		return err
	}
	if nav.ErrorText != "" {
		return errors.New(nav.ErrorText)
	}
	if nav.LoaderID == "" {
		close(eventFired)
	}
	select {
	case <-eventFired:
	case <-time.After(NavigationTimeout):
		return errors.New("navigate '" + urlStr + "' reached timeout")
	}
	session.switchContext(nav.FrameID)
	return nil
}

// Reload refresh current page ignores cache
func (session *Session) Reload() error {
	eventFired := make(chan bool)
	unsubscribe := session.Subscribe("Page.domContentEventFired", func(params Params) {
		eventFired <- true
	})
	defer unsubscribe()
	if err := session.reload(); err != nil {
		return err
	}
	select {
	case <-eventFired:
	case <-time.After(NavigationTimeout):
		return errors.New("reload reached timeout")
	}
	session.MainFrame()
	return nil
}

// Close close this sessions
func (session *Session) Close() (bool, error) {
	return session.closeTarget(session.targetID)
}

// GetScreenshot get screen of current page
func (session *Session) GetScreenshot(format ImageFormat, quality int8, clip *Viewport, fullPage bool) ([]byte, error) {
	if fullPage {
		view, err := session.getLayoutMetrics()
		if err != nil {
			return nil, err
		}
		defer session.clearDeviceMetricsOverride()
		session.setDeviceMetricsOverride(int64(math.Ceil(view.ContentSize.Width)), int64(math.Ceil(view.ContentSize.Height)), 1)
	}
	return session.captureScreenshot(format, quality, clip)
}

// TargetCreated subscribe to targetCreated event
func (session *Session) TargetCreated() chan *TargetInfo {
	message := make(chan *TargetInfo)
	var unsubscribe func()
	unsubscribe = session.Subscribe("Target.targetCreated", func(msg Params) {
		targetInfo := &TargetInfo{}
		unmarshal(msg["targetInfo"], targetInfo)
		if targetInfo.Type == "page" {
			message <- targetInfo
			unsubscribe()
		}
	})
	return message
}

// IsClosed is session closed?
func (session *Session) IsClosed() bool {
	select {
	case <-session.closed:
		return true
	default:
		return false
	}
}

// Script evaluate javascript code at context of web page synchronously
func (session *Session) Script(code string) (interface{}, error) {
	result, err := session.Evaluate(code, 0)
	if err != nil {
		return "", err
	}
	return result.Value, nil
}

// CallScriptOn evaluate javascript code on element
// executed as function() { `script` }, for access to element use `this`
// for example script = `this.innerText = "test"`
func (session *Session) CallScriptOn(selector, script string) (interface{}, error) {
	element, err := session.findElement(selector)
	if err != nil {
		return nil, err
	}
	defer session.release(element)
	result, err := session.callFunctionOn(element, `function(){`+script+`}`)
	if err != nil {
		return "", err
	}
	return result.Value, nil
}

// URL get current tab's url
func (session *Session) URL() (string, error) {
	history, err := session.getNavigationHistory()
	if err != nil {
		return "", err
	}
	if history.CurrentIndex == -1 {
		return "about:blank", nil
	}
	return history.Entries[history.CurrentIndex].URL, nil
}

// Title get current tab's title
func (session *Session) Title() (string, error) {
	history, err := session.getNavigationHistory()
	if err != nil || history.CurrentIndex == -1 {
		return "", err
	}
	return history.Entries[history.CurrentIndex].Title, nil
}

func (session *Session) listener() {
	var method string
	for e := range session.incomingEvent {

		method = e["method"].(string)
		session.rw.RLock()
		cbs, has := session.callbacks[method]
		session.rw.RUnlock()
		if has {
			for _, c := range cbs {
				if c != nil {
					c(e["params"].(Params))
				}
			}
		}

		switch method {
		case "Runtime.executionContextsCleared":
			session.contextID = 0

		case "Runtime.executionContextCreated":
			desc := &ExecutionContextDescription{}
			unmarshal(e["params"].(map[string]interface{})["context"], desc)
			frameID := desc.AuxData["frameId"].(string)
			if session.frameID == frameID {
				session.contextID = desc.ID
			}

		case "Runtime.executionContextDestroyed":
			executionContextID := e["params"].(map[string]interface{})["executionContextId"]
			if session.contextID == executionContextID {
				session.contextID = 0
			}

		case "Target.targetCrashed":
			crashed := &targetCrashed{}
			unmarshal(e["params"], crashed)
			panic(crashed.Status)

		case "Target.targetDestroyed":
			targetID := e["params"].(Params)["targetId"].(string)
			if targetID == session.targetID {
				close(session.closed)
				session.client.deleteSession(session.sessionID)
				return
			}
		}
	}
}

func (session *Session) blockingSend(method string, params *Params) (MessageResult, error) {
	recv := session.client.sendMethod(session.sessionID, method, params)
	select {
	case resp := <-recv:
		if obj, has := resp["error"]; has {
			mes := obj.(Params)["message"].(string)
			return nil, errors.New(mes)
		}
		return resp["result"].(MessageResult), nil
	case <-session.closed:
		return nil, ErrSessionClosed
	case <-time.After(WebSocketTimeout):
		panic(fmt.Sprintf("websocket response reached timeout %s for %s -> %+v", WebSocketTimeout.String(), method, params))
	}
}

// Subscribe subscribe to CDP event
func (session *Session) Subscribe(method string, callback func(params Params)) (unsubscribe func()) {
	session.rw.Lock()
	if _, has := session.callbacks[method]; !has {
		session.callbacks[method] = make([]func(Params), 0)
	}
	session.callbacks[method] = append(session.callbacks[method], callback)
	index := len(session.callbacks[method]) - 1
	session.rw.Unlock()

	return func() {
		session.rw.Lock()
		session.callbacks[method][index] = nil
		session.rw.Unlock()
	}
}
