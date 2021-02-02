package cdp

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"math"
	"sync"
	"time"

	"github.com/ecwid/cdp/pkg/devtool"
)

// Network domain
type Network = Session

// DOM domain
type DOM = Session

// Memory domain
type Memory = Session

// Input domain
type Input = Session

// Runtime domain
type Runtime = Session

// Emulation domain
type Emulation = Session

// Page domain
type Page = Session

func (session Session) loadEventFired() func() error {
	c := make(chan struct{})
	unsubscribe := session.Subscribe("Page.loadEventFired", func(*Event) {
		select {
		case c <- struct{}{}:
		default:
		}
	})
	return func() error {
		defer close(c)
		defer unsubscribe()
		select {
		case <-c:
			return nil
		case <-session.closed:
			return ErrSessionAlreadyClosed
		case <-time.After(session.deadline):
			return ErrLoadTimeout
		}
	}
}

// Navigate navigate to url
func (session Session) Navigate(urlStr string) (err error) {
	loader := session.loadEventFired()
	nav := new(devtool.NavigationResult)
	p := Map{
		"url":            urlStr,
		"transitionType": "typed",
		"frameId":        session.target,
	}
	if err = session.call("Page.navigate", p, nav); err != nil {
		return err
	}
	if nav.ErrorText != "" {
		return errors.New(nav.ErrorText)
	}
	if nav.LoaderID == "" {
		return nil // no navigate need
	}
	return loader()
}

// Reload refresh current page ignores cache
func (session Session) Reload() error {
	loader := session.loadEventFired()
	if err := session.call("Page.reload", Map{"ignoreCache": true}, nil); err != nil {
		return err
	}
	session.state.reset()
	return loader()
}

// OnTargetCreated subscribe to Target.targetCreated event and return channel with targetID
func (session Session) OnTargetCreated(before func()) (*Session, error) {
	var eventFired = make(chan string, 1)
	unsubscribe := session.Subscribe("Target.targetCreated", func(e *Event) {
		targetCreated := new(devtool.TargetCreated)
		if err := json.Unmarshal(e.Params, targetCreated); err != nil {
			session.exception(err)
			return
		}
		if targetCreated.TargetInfo.Type == "page" && targetCreated.TargetInfo.OpenerID == session.target {
			select {
			case eventFired <- targetCreated.TargetInfo.TargetID:
			default:
			}
		}
	})
	defer close(eventFired)
	defer unsubscribe()
	before()
	select {
	case id := <-eventFired:
		return NewSession(&session, id)
	case err := <-session.err:
		return nil, err
	case <-session.closed:
		return nil, ErrSessionAlreadyClosed
	case <-time.After(session.deadline):
		return nil, ErrTargetCreatedTimeout
	}
}

// Main switch context to main frame of page
func (session Session) Main() {
	session.state.reset()
}

// SwitchTo switch context to frame
func (session *Session) SwitchTo(frameID string) error {
	c, err := session.createContext(frameID)
	if err != nil {
		return err
	}
	session.state.set(frameID, c)
	return nil
}

// Activate activate current Target
func (session Session) Activate() error {
	return session.activate(session.target)
}

// NewTab ...
func (session Session) NewTab(url string) (*Session, error) {
	if url == "" {
		url = blankPage // headless chrome crash when url is empty
	}
	result := Map{}
	if err := session.call("Target.createTarget", Map{"url": url}, &result); err != nil {
		return nil, err
	}
	return NewSession(&session, result["targetId"].(string))
}

// Query query element on page by css selector
func (session Session) Query(selector string) (*Element, error) {
	return session.query(nil, selector)
}

// QueryAll queryAll elements on page by css selector
func (session Session) QueryAll(selector string) ([]*Element, error) {
	return session.queryAll(nil, selector)
}

// NavigateHistory -1 = Back, +1 = Forward
func (session Session) NavigateHistory(delta int64) error {
	history, err := session.getNavigationHistory()
	if err != nil {
		return err
	}
	move := history.CurrentIndex + delta
	if move >= 0 && move < int64(len(history.Entries)) {
		return session.navigateToHistoryEntry(history.Entries[move].ID)
	}
	return nil
}

// GetNavigationEntry get current tab info
func (session Session) GetNavigationEntry() (*devtool.NavigationEntry, error) {
	history, err := session.getNavigationHistory()
	if err != nil {
		return nil, err
	}
	if history.CurrentIndex == -1 {
		return &devtool.NavigationEntry{URL: blankPage}, nil
	}
	return history.Entries[history.CurrentIndex], nil
}

// FitToWindow ...
func (session Session) FitToWindow() error {
	view, err := session.getLayoutMetrics()
	if err != nil {
		return err
	}
	return session.SetDeviceMetricsOverride(&devtool.DeviceMetrics{
		Width:             view.LayoutViewport.ClientWidth,
		Height:            int64(math.Ceil(view.ContentSize.Height)),
		DeviceScaleFactor: 1,
		Mobile:            false,
	})
}

// CaptureScreenshot get screen of current page
func (session Session) CaptureScreenshot(format string, quality int8) ([]byte, error) {
	if err := session.SetScrollbarsHidden(true); err != nil {
		return nil, err
	}
	p := Map{
		"format":      format,
		"quality":     quality,
		"fromSurface": true,
	}
	result := Map{}
	err := session.call("Page.captureScreenshot", p, &result)
	if err != nil {
		return nil, err
	}
	return base64.StdEncoding.DecodeString(result["data"].(string))
}

// Listen subscribe to listen cdp events with methods name
// return channel with incoming events and func to unsubscribe
// channel will be closed after unsubscribe func call
func (session Session) Listen(methods ...string) (chan *Event, func()) {
	var (
		wg          = sync.WaitGroup{}
		queue       = make(chan *Event, 10)
		interrupt   = make(chan struct{})
		unsubscribe = make([]func(), len(methods))
	)
	callback := func(e *Event) {
		wg.Add(1)
		defer wg.Done()
		select {
		case queue <- e:
		case <-interrupt:
		}
	}
	for n, m := range methods {
		unsubscribe[n] = session.Subscribe(m, callback)
	}
	return queue, func() {
		close(interrupt)
		for _, un := range unsubscribe {
			un()
		}
		wg.Wait()
		close(queue)
	}
}
