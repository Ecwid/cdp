package cdp

import (
	"encoding/json"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/ecwid/cdp/har"
)

// NavigationTimeout - navigation timeout
var NavigationTimeout = time.Second * 40

// NetworkTimeout - network timeout
var NetworkTimeout = time.Second * 10

// WebSocketTimeout - web socket response timeout
var WebSocketTimeout = time.Second * 100

// ErrSessionClosed current session (target) already closed
var ErrSessionClosed = errors.New(`Session closed`)

// Session CDP сессия
type Session struct {
	rw            sync.RWMutex
	client        *Client
	sessionID     string
	targetID      string
	contextID     int64
	frameID       string
	incomingEvent chan MessageResult // очередь событий на обработку от websocket клиента
	callbacks     map[string]func(Params)
	closed        chan bool
	har           *HAR
}

func newSession(client *Client) *Session {
	session := &Session{
		client:        client,
		incomingEvent: make(chan MessageResult, 2000),
		callbacks:     make(map[string]func(Params)),
		closed:        make(chan bool, 1),
	}
	go session.listener()
	return session
}

// Client ...
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

// SwitchToFrame переключается на фрейм элемента selector
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

// MainFrame переключается на главный фрейм страницы
func (session *Session) MainFrame() {
	session.switchContext(session.targetID)
}

// Navigate ...
func (session *Session) Navigate(urlStr string) error {
	eventFired := make(chan bool)
	unsubscribe := session.Subscribe("Page.loadEventFired", func(params Params) {
		eventFired <- true
	})
	defer unsubscribe()
	nav, err := session.navigate(urlStr, session.frameID)
	if err != nil {
		panic(err)
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
		panic(err)
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
func (session *Session) GetScreenshot(format string, quality int8, fullPage bool) ([]byte, error) {
	if fullPage {
		lm, err := session.getLayoutMetrics()
		if err != nil {
			return nil, err
		}
		defer session.clearDeviceMetricsOverride()
		session.setDeviceMetricsOverride(int64(math.Ceil(lm.ContentSize.Width)), int64(math.Ceil(lm.ContentSize.Height)), 1)
	}
	raw, err := session.captureScreenshot(format, quality)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

// TargetCreated ...
func (session *Session) TargetCreated() chan *TargetInfo {
	message := make(chan *TargetInfo)
	session.Subscribe("Target.targetCreated", func(msg Params) {
		targetInfo := &TargetInfo{}
		unmarshal(msg["targetInfo"], targetInfo)
		if targetInfo.Type == "page" {
			message <- targetInfo
			delete(session.callbacks, "Target.targetCreated")
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

// Script evaluate javascript code at context of web page
func (session *Session) Script(code string) (interface{}, error) {
	result, err := session.Evaluate(code, 0)
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

// HARBegin start HAR records
func (session *Session) HARBegin() {
	session.networkEnable()
	session.har = &HAR{
		Log: &har.HAR{
			Version: "1.2",
			Creator: &har.Creator{
				Name:    "ecwid-cdp",
				Version: "0.1",
			},
			Pages:   make([]*har.Page, 0),
			Entries: make([]*har.Entry, 0),
		},
	}
}

// GetHARRequest ...
func (session *Session) GetHARRequest(requestID string) *har.Request {
	if session.har == nil {
		return nil
	}
	e := session.har.entryByRequestID(requestID)
	if e == nil {
		return nil
	}
	return e.Request
}

// HAREnd stop and get recorded HAR
func (session *Session) HAREnd() *HAR {
	har := session.har
	session.har = nil
	return har
}

func (session *Session) listener() {
	var method string
	for e := range session.incomingEvent {

		method = e["method"].(string)
		session.rw.RLock()
		cb, has := session.callbacks[method]
		session.rw.RUnlock()
		if has {
			cb(e["params"].(Params))
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
		case "Page.loadEventFired",
			"Page.domContentEventFired",
			"Page.frameStartedLoading",
			"Page.frameAttached",
			"Page.navigatedWithinDocument",
			"Network.requestWillBeSent",
			"Network.requestServedFromCache",
			"Network.dataReceived",
			"Network.responseReceived",
			"Network.resourceChangedPriority",
			"Network.loadingFinished",
			"Network.loadingFailed":
			if session.har != nil {
				if err := session.eventHar(method, e["params"]); err != nil {
					// log.Printf(err.Error())
				}
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
		panic(`websocket message response reached timeout`)
	}
}

// Subscribe subscribe to CDP event
func (session *Session) Subscribe(method string, callback func(params Params)) (unsubscribe func()) {
	session.rw.RLock()
	_, has := session.callbacks[method]
	session.rw.RUnlock()
	if has {
		panic(`listener for event ` + method + ` already exist`)
	}
	session.rw.Lock()
	session.callbacks[method] = callback
	session.rw.Unlock()
	return func() {
		session.rw.Lock()
		delete(session.callbacks, method)
		session.rw.Unlock()
	}
}
