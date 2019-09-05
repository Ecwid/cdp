package agent

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/ecwid/cdp"
)

// New create new agent
func New(client *cdp.Client) *Agent {
	agent := &Agent{
		stack:    make([]*cdp.Session, 1),
		Deadline: time.Second * 60,
		Delay:    time.Millisecond * 500,
	}
	agent.stack[0] = client.NewSession(nil)
	return agent
}

// Agent ...
type Agent struct {
	stack    []*cdp.Session
	Deadline time.Duration
	Delay    time.Duration
	Logger   func(string, []interface{}, []interface{})
}

func (a *Agent) active() *cdp.Session {
	return a.stack[len(a.stack)-1]
}

// Session get current session
func (a *Agent) Session() *cdp.Session {
	return a.active()
}

func (a *Agent) log(params []interface{}, result []interface{}) {
	if a.Logger == nil {
		return
	}
	pc, _, _, _ := runtime.Caller(3)
	name := runtime.FuncForPC(pc).Name()
	name = name[strings.LastIndex(name, ".")+1 : len(name)]
	a.Logger(name, params, result)
}

func (a *Agent) proxy(fn func(s *cdp.Session) error) {
	var err error
	for start := time.Now(); time.Since(start) < a.Deadline; {
		if err = fn(a.active()); err == nil {
			return
		}
		time.Sleep(a.Delay)
	}
	if err == nil {
		panic("agent timeout")
	}
	panic(err)
}

func (a *Agent) proxy0(fn func(s *cdp.Session) error) {
	var err error
	if err = fn(a.active()); err == nil {
		return
	}
	if err == nil {
		panic("agent timeout")
	}
	panic(err)
}

// CurrentURL get current page URL
func (a *Agent) CurrentURL() string {
	var url string
	var e error
	a.proxy(func(s *cdp.Session) error {
		url, e = s.URL()
		a.log(nil, []interface{}{url, e})
		return e
	})
	return url
}

// Close close current target (session closes as well and can't be used anymore)
func (a *Agent) Close() {
	a.proxy0(func(s *cdp.Session) error {
		_, e := s.Close()
		a.log(nil, []interface{}{e})
		return e
	})
}

// GetScreenshot capture screen of current page with `format` and `quality`
func (a *Agent) GetScreenshot(format cdp.ImageFormat, quality int8, full bool) []byte {
	var img []byte
	var e error
	a.proxy0(func(s *cdp.Session) error {
		img, e = s.GetScreenshot(format, quality, nil, full)
		a.log(nil, []interface{}{e})
		return nil
	})
	return img
}

// MainFrame switch context to main frame
func (a *Agent) MainFrame() {
	a.proxy0(func(s *cdp.Session) error {
		s.MainFrame()
		a.log(nil, nil)
		return nil
	})
}

// Reload reload current page
func (a *Agent) Reload() {
	a.proxy(func(s *cdp.Session) error {
		e := s.Reload()
		a.log(nil, []interface{}{e})
		return e
	})
}

// SetHeaders set extra headers
func (a *Agent) SetHeaders(headers map[string]string) {
	a.proxy0(func(s *cdp.Session) error {
		s.SetExtraHTTPHeaders(headers)
		a.log([]interface{}{headers}, nil)
		return nil
	})
}

// Script call javascript code in main execution context
func (a *Agent) Script(code string) (interface{}, error) {
	var r interface{}
	var e error
	a.proxy0(func(s *cdp.Session) error {
		r, e = s.Script(code)
		a.log([]interface{}{code}, []interface{}{r, e})
		return nil
	})
	return r, e
}

// Navigate navigate page to address
func (a *Agent) Navigate(url string) {
	a.proxy(func(s *cdp.Session) error {
		e := s.Navigate(url)
		a.log([]interface{}{url}, []interface{}{e})
		return e
	})
}

// SetCookies set browser's cookies
func (a *Agent) SetCookies(cookies ...cdp.CookieParam) {
	a.proxy0(func(s *cdp.Session) error {
		s.SetCookies(cookies...)
		a.log([]interface{}{cookies}, nil)
		return nil
	})
}

// SwitchToFrame switch to frame by frame element selector
func (a *Agent) SwitchToFrame(selector string) {
	a.proxy(func(s *cdp.Session) error {
		e := s.SwitchToFrame(selector)
		a.log([]interface{}{selector}, []interface{}{e})
		return e
	})
}

// GetComputedStyle get computed style for element `selector`
func (a *Agent) GetComputedStyle(selector string, style string) string {
	var comp = ""
	var err error
	a.proxy(func(s *cdp.Session) error {
		comp, err = s.GetComputedStyle(selector, style)
		a.log([]interface{}{selector, style}, []interface{}{comp, err})
		return err
	})
	return comp
}

// Select select options by values (multiple select supported)
func (a *Agent) Select(selector string, args ...string) {
	a.proxy(func(s *cdp.Session) error {
		e := s.Select(selector, args...)
		a.log([]interface{}{selector, args}, []interface{}{e})
		return e
	})
}

// Upload upload file in input
func (a *Agent) Upload(selector string, file string) {
	a.proxy(func(s *cdp.Session) error {
		e := s.Upload(selector, file)
		a.log([]interface{}{selector, file}, []interface{}{e})
		return e
	})
}

// getSelected get selected options, selectedText - get option's text or value
func (a *Agent) getSelected(selector string, textElseValue bool) []string {
	var text []string
	var err error
	a.proxy(func(s *cdp.Session) error {
		text, err = s.GetSelected(selector, textElseValue)
		a.log([]interface{}{selector, textElseValue}, []interface{}{text, err})
		return err
	})
	return text
}

// GetSelected получает value выбранной опции из select
func (a *Agent) GetSelected(selector string) []string {
	return a.getSelected(selector, false)
}

// GetSelectedText получает текст выбранной опции из select
func (a *Agent) GetSelectedText(selector string) []string {
	return a.getSelected(selector, true)
}

// Click click at visible selector
func (a *Agent) Click(selector string) {
	a.proxy(func(s *cdp.Session) error {
		e := s.Click(selector)
		a.log([]interface{}{selector}, []interface{}{e})
		return e
	})
}

// Checkbox set checkbox
func (a *Agent) Checkbox(selector string, check bool) {
	a.proxy(func(s *cdp.Session) error {
		e := s.Checkbox(selector, check)
		a.log([]interface{}{selector, check}, []interface{}{e})
		return e
	})
}

// IsChecked is checkbox checked
func (a *Agent) IsChecked(selector string) bool {
	var checked bool
	var err error
	a.proxy(func(s *cdp.Session) error {
		checked, err = s.IsChecked(selector)
		a.log([]interface{}{selector}, []interface{}{checked, err})
		return err
	})
	return checked
}

// SetAttr set element attribute
func (a *Agent) SetAttr(selector string, attr string, value string) {
	a.proxy(func(s *cdp.Session) error {
		e := s.SetAttr(selector, attr, value)
		a.log([]interface{}{selector, attr, value}, []interface{}{e})
		return e
	})
}

// GetAttr get element attribute
func (a *Agent) GetAttr(selector string, attr string) string {
	var value string
	var err error
	a.proxy(func(s *cdp.Session) error {
		value, err = s.GetAttr(selector, attr)
		a.log([]interface{}{selector, attr}, []interface{}{value, err})
		return err
	})
	return value
}

// Type insert text into element and press keys
func (a *Agent) Type(selector string, text string, keys ...rune) {
	a.proxy(func(s *cdp.Session) error {
		e := s.Type(selector, text, keys...)
		a.log([]interface{}{selector, text, keys}, []interface{}{e})
		if e != nil {
			return e
		}
		var gettext []string
		gettext, e = s.Text(selector)
		if e != nil {
			return e
		}
		if text != gettext[0] {
			return fmt.Errorf("`%s` != `%s`", text, gettext[0])
		}
		return nil
	})
}

// Predicate waiting for predicate be true
func (a *Agent) Predicate(text string, predicate func() bool) {
	a.proxy(func(s *cdp.Session) error {
		if predicate() {
			return nil
		}
		return fmt.Errorf(text)
	})
}

// Exist waiting for appear selector
func (a *Agent) Exist(selector string) {
	a.Predicate(fmt.Sprintf("%s not appears", selector), func() bool {
		return a.Count(selector) > 0
	})
}

// NotExist waiting for disappear selector
func (a *Agent) NotExist(selector string) {
	a.Predicate(fmt.Sprintf("%s not disappears", selector), func() bool {
		return a.Count(selector) == 0
	})
}

// SelectExist waiting for one of selectors appears. At least one element should be appers. First appeared will be returned.
func (a *Agent) SelectExist(selector ...string) string {
	m := fmt.Sprintf("no one of [%s] exist", strings.Join(selector, " or "))
	var exist = ""
	a.Predicate(m, func() bool {
		for _, s := range selector {
			if a.Count(s) > 0 {
				exist = s
				return true
			}
		}
		return false
	})
	return exist
}

// Displayed waiting for selector displayed
func (a *Agent) Displayed(selector string) {
	a.proxy(func(s *cdp.Session) error {
		v, e := s.IsDisplayed(selector)
		a.log([]interface{}{selector}, []interface{}{v, e})
		if e != nil {
			return e
		}
		if !v {
			return errors.New("element not displayed")
		}
		return nil
	})
}

// IsClosed check if tab was closed
func (a *Agent) IsClosed() bool {
	var closed bool
	a.proxy0(func(s *cdp.Session) error {
		closed = s.IsClosed()
		a.log([]interface{}{}, []interface{}{closed})
		return nil
	})
	return closed
}

// Closed waiting for target closes
func (a *Agent) Closed() {
	a.Predicate("target "+a.active().ID()+" not closed", func() bool {
		return a.IsClosed()
	})
}

// Text get element's text (innerText or input's value)
func (a *Agent) Text(selector string) string {
	texts := a.TextAll(selector)
	if texts != nil && len(texts) > 0 {
		return texts[0]
	}
	return ""
}

// TextAll get elements text query by selector
func (a *Agent) TextAll(selector string) []string {
	var text []string
	var err error
	a.proxy(func(s *cdp.Session) error {
		text, err = s.Text(selector)
		a.log([]interface{}{selector}, []interface{}{text, err})
		return err
	})
	return text
}

// GetRectangle get dimension of element
func (a *Agent) GetRectangle(selector string) *cdp.Rect {
	var rect *cdp.Rect
	var err error
	a.proxy(func(s *cdp.Session) error {
		rect, err = s.GetRectangle(selector)
		a.log([]interface{}{selector}, []interface{}{rect, err})
		return err
	})
	return rect
}

// Hover mouse hover on element
func (a *Agent) Hover(selector string) {
	a.proxy(func(s *cdp.Session) error {
		e := s.Hover(selector)
		a.log([]interface{}{selector}, []interface{}{e})
		return e
	})
}

// CallScriptOn evaluate javascript code on element
// executed as function() { `script` }, for access to element use `this`
// for example script = `this.innerText = "test"`
func (a *Agent) CallScriptOn(selector, script string) interface{} {
	var result interface{}
	var err error
	a.proxy(func(s *cdp.Session) error {
		result, err = s.CallScriptOn(selector, script)
		a.log([]interface{}{selector, script}, []interface{}{result, err})
		return err
	})
	return result
}

// SendKeys send keyboard keys to focused element
func (a *Agent) SendKeys(key ...rune) {
	a.proxy0(func(s *cdp.Session) error {
		e := s.SendKeys(key...)
		a.log([]interface{}{key}, []interface{}{e})
		return e
	})
}

// NewTarget call init and waiting for new target created
func (a *Agent) NewTarget(init func()) string {
	var targetInfo chan *cdp.TargetInfo
	a.proxy0(func(s *cdp.Session) error {
		targetInfo = s.TargetCreated()
		a.log(nil, nil)
		return nil
	})
	init()
	select {
	case target := <-targetInfo:
		return target.TargetID
	case <-time.After(cdp.NavigationTimeout):
		panic("new target has not been created")
	}
}

// Count count of elements in DOM
func (a *Agent) Count(selector string) int {
	var count = 0
	a.proxy0(func(s *cdp.Session) error {
		count = s.Count(selector)
		a.log([]interface{}{selector}, []interface{}{count})
		return nil
	})
	return count
}

// NewTab call do() func to a new tab opened then switch to the new tab
func (a *Agent) NewTab(do func()) {
	target := a.NewTarget(do)
	sess := a.active().Client().NewSession(&target)
	a.stack = append(a.stack, sess)
}

// PrevTab switch to prev tab in history
func (a *Agent) PrevTab() {
	if len(a.stack) == 0 {
		return
	}
	n := len(a.stack) - 1 // Top element
	a.stack = a.stack[:n] // Pop
}

// SetInnerText set innerText for node
func (a *Agent) SetInnerText(selector string, text string) {
	script := fmt.Sprintf(`this.innerText="%s"`, text)
	a.CallScriptOn(selector, script)
}

// RemoveElement remove node from DOM tree
func (a *Agent) RemoveElement(selector string) {
	a.CallScriptOn(selector, `this.remove()`)
}

// Scroll scroll to coords
func (a *Agent) Scroll(x, y int) {
	if _, e := a.Script(fmt.Sprintf(`scrollTo(%d, %d)`, x, y)); e != nil {
		panic(e)
	}
}

// SetLocalStorage устанавливаем localStorage в браузере
func (a *Agent) SetLocalStorage(name, value string) {
	if _, e := a.Script(fmt.Sprintf(`localStorage.setItem("%s", %s)`, name, value)); e != nil {
		panic(e)
	}
}
