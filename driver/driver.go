package driver

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/ecwid/cdp"
)

// DriverMaxRetry specifies the amount of time the driver should wait when call browser actions
var DriverMaxRetry = time.Second * 60

// DriverRetryDelay specifies delay between retries
var DriverRetryDelay = time.Millisecond * 500

// Screenshot image types
const (
	PNG  = "png"
	JPEG = "jpeg"
)

// Cookie alias for CDP Cookie
type Cookie = cdp.CookieParam

// Driver error tolerance wrapper for CDP session
type Driver struct {
	PanicInterceptor func(error)
	Logger           func(string, []interface{}, []interface{})
	session          *cdp.Session
}

// Session current session associated with Driver
func (e *Driver) Session() *cdp.Session {
	return e.session
}

// NewDriver open new target and raise session for it
func NewDriver(client *cdp.Client) *Driver {
	return &Driver{session: client.NewSession(nil)}
}

func (e *Driver) log(params []interface{}, result []interface{}) {
	if e.Logger == nil {
		return
	}
	pc, _, _, _ := runtime.Caller(1)
	name := runtime.FuncForPC(pc).Name()
	name = name[strings.LastIndex(name, ".")+1 : len(name)]
	e.Logger(name, params, result)
}

// Panic interceptor for package panic. Uses native panic if PanicInterceptor == nil
func (e *Driver) Panic(cause error, format string, v ...interface{}) {
	var err = &ErrDriver{
		SessionID: e.session.ID(),
		Message:   fmt.Sprintf(format, v...),
		Cause:     cause,
	}
	if e.PanicInterceptor != nil {
		e.PanicInterceptor(err)
	} else {
		panic(err)
	}
}

func retry(retry func() error) error {
	backoff := &backoff{
		Max:   DriverMaxRetry,
		Delay: DriverRetryDelay,
		Retry: retry,
	}
	return backoff.Do()
}

// CurrentURL get current page URL
func (e *Driver) CurrentURL() string {
	var url string
	var err error
	_ = retry(func() error {
		url, err = e.session.URL()
		return err
	})
	e.log(nil, []interface{}{url})
	if err != nil {
		e.Panic(err, "get current page url failed")
	}
	return url
}

// Close close current target (session closes as well and can't be used anymore)
func (e *Driver) Close() {
	_, err := e.session.Close()
	e.log(nil, nil)
	if err != nil {
		e.Panic(err, "close current target failed")
	}
}

// GetScreenshot capture screen of current page with `format` and `quality`
func (e *Driver) GetScreenshot(format string, quality int8, full bool) []byte {
	img, _ := e.session.GetScreenshot(format, quality, full)
	return img
}

// MainFrame switch context to main frame
func (e *Driver) MainFrame() {
	e.session.MainFrame()
	e.log(nil, nil)
}

// Reload reload current page
func (e *Driver) Reload() {
	err := retry(func() error {
		return e.session.Reload()
	})
	e.log(nil, nil)
	if err != nil {
		e.Panic(err, "reload page failed")
	}
}

// SetHeaders set extra headers
func (e *Driver) SetHeaders(headers map[string]string) {
	e.session.SetExtraHTTPHeaders(headers)
	e.log([]interface{}{headers}, nil)
}

// Script call javascript code in main execution context
func (e *Driver) Script(code string) (interface{}, error) {
	r, err := e.session.Script(code)
	e.log([]interface{}{code}, []interface{}{r, err})
	return r, err
}

// Navigate navigate page to address
func (e *Driver) Navigate(url string) {
	err := retry(func() error {
		return e.session.Navigate(url)
	})
	e.log([]interface{}{url}, nil)
	if err != nil {
		e.Panic(err, "navigate to `%s` failed", url)
	}
}

// SetCookies set browser's cookies
func (e *Driver) SetCookies(cookies ...Cookie) {
	e.session.SetCookies(cookies...)
	e.log([]interface{}{cookies}, nil)
}

// SwitchToFrame switch to frame by frame element selector
func (e *Driver) SwitchToFrame(selector string) {
	err := retry(func() error {
		return e.session.SwitchToFrame(selector)
	})
	e.log([]interface{}{selector}, nil)
	if err != nil {
		e.Panic(err, "switch to frame `%s` failed", selector)
	}
}

// NewDriver new driver by target ID
func (e *Driver) NewDriver(targetID string) *Driver {
	// Создаем новую сессию драйвера
	driver := &Driver{
		PanicInterceptor: e.PanicInterceptor,
		Logger:           e.Logger,
		session:          e.session.Client().NewSession(&targetID),
	}
	e.log([]interface{}{targetID}, nil)
	return driver
}

// GetComputedStyle get computed style for element `selector`
func (e *Driver) GetComputedStyle(selector string, style string) string {
	var comp = ""
	var err error
	_ = retry(func() error {
		comp, err = e.session.GetComputedStyle(selector, style)
		return err
	})
	e.log([]interface{}{selector, style}, []interface{}{comp})
	if err != nil {
		e.Panic(err, "get computed style [%s] for selector `%s` failed", style, selector)
	}
	return comp
}

// Select select options by values (multiple select supported)
func (e *Driver) Select(selector string, args ...string) {
	err := retry(func() error {
		return e.session.Select(selector, args...)
	})
	e.log([]interface{}{selector, args}, nil)
	if err != nil {
		e.Panic(err, "select option(s) %v for selector %s failed", args, selector)
	}
}

// Upload upload file in input
func (e *Driver) Upload(selector string, file string) {
	err := retry(func() error {
		return e.session.Upload(selector, file)
	})
	e.log([]interface{}{selector, file}, nil)
	if err != nil {
		e.Panic(err, "upload file for selector `%s` failed", selector)
	}
}

// getSelected get selected options, selectedText - get option's text or value
func (e *Driver) getSelected(selector string, selectedText bool) []string {
	var text []string
	var err error
	_ = retry(func() error {
		text, err = e.session.GetSelected(selector, selectedText)
		return err
	})
	e.log([]interface{}{selector, selectedText}, []interface{}{text})
	if err != nil {
		e.Panic(err, "get selected options for selector `%s` failed", selector)
	}
	return text
}

// GetSelected получает value выбранной опции из select
func (e *Driver) GetSelected(selector string) []string {
	return e.getSelected(selector, false)
}

// GetSelectedText получает текст выбранной опции из select
func (e *Driver) GetSelectedText(selector string) []string {
	return e.getSelected(selector, true)
}

// Click click at visible selector
func (e *Driver) Click(selector string) {
	err := retry(func() error {
		return e.session.Click(selector)
	})
	e.log([]interface{}{selector}, nil)
	if err != nil {
		e.Panic(err, "click at selector `%s` failed", selector)
	}
}

// Checkbox set checkbox
func (e *Driver) Checkbox(selector string, check bool) {
	err := retry(func() error {
		return e.session.Checkbox(selector, check)
	})
	e.log([]interface{}{selector, check}, nil)
	if err != nil {
		e.Panic(err, "set checkbox [%t] for selector %s failed", check, selector)
	}
}

// IsChecked is checkbox checked
func (e *Driver) IsChecked(selector string) bool {
	var checked bool
	var err error
	_ = retry(func() error {
		checked, err = e.session.IsChecked(selector)
		return err
	})
	e.log([]interface{}{selector}, []interface{}{checked})
	if err != nil {
		e.Panic(err, "get checkbox `%s` value failed", selector)
	}
	return checked
}

// SetAttr set element attribute
func (e *Driver) SetAttr(selector string, attr string, value string) {
	err := retry(func() error {
		return e.session.SetAttr(selector, attr, value)
	})
	e.log([]interface{}{selector, attr, value}, nil)
	if err != nil {
		e.Panic(err, "set attribute [%s = %s] for element `%s` failed", attr, value, selector)
	}
}

// GetAttr get element attribute
func (e *Driver) GetAttr(selector string, attr string) string {
	var value string
	var err error
	_ = retry(func() error {
		value, err = e.session.GetAttr(selector, attr)
		return err
	})
	e.log([]interface{}{selector, attr}, []interface{}{value})
	if err != nil {
		e.Panic(err, "get attribute [%s] for element `%s` failed", attr, selector)
	}
	return value
}

// Type insert text into element and press keys
func (e *Driver) Type(selector string, text string, keys ...rune) {
	err := retry(func() error {
		err := e.session.Type(selector, text, keys...)
		if err != nil {
			return err
		}
		var gettext []string
		gettext, err = e.session.Text(selector)
		if err != nil {
			return err
		}
		if text != gettext[0] {
			return fmt.Errorf("not expected text was typed exp: `%s`, was: `%s`", text, gettext[0])
		}
		return nil
	})
	e.log([]interface{}{selector, text, keys}, nil)
	if err != nil {
		e.Panic(err, "type text for selector `%s` failed", selector)
	}
}

// Predicate waiting for predicate be true
func (e *Driver) Predicate(text string, predicate func() bool) {
	err := retry(func() error {
		if predicate() {
			return nil
		}
		return errors.New("not satisfied")
	})
	e.log([]interface{}{text}, nil)
	if err != nil {
		e.Panic(nil, "expected condition `%s` not satisfied", text)
	}
}

// Exist waiting for one of selectors appears. At least one element should be appers. First appeared will be returned.
// Please use this function instead of selenium's isPresence
func (e *Driver) Exist(selector ...string) string {
	m := fmt.Sprintf("[%s] have to appears", strings.Join(selector, " or "))
	var exist = ""
	e.Predicate(m, func() bool {
		for _, s := range selector {
			if e.session.Count(s) > 0 {
				exist = s
				return true
			}
		}
		return false
	})
	e.log([]interface{}{selector}, []interface{}{exist})
	return exist
}

// Displayed waiting for selector displayed
func (e *Driver) Displayed(selector string) {
	err := retry(func() error {
		v, err := e.session.IsDisplayed(selector)
		if err != nil {
			return err
		}
		if v {
			return nil
		}
		return errors.New("element not displayed")
	})
	if err != nil {
		e.Panic(err, "element `%s` not displayed", selector)
	}
}

// NotExist waiting for disappear selector
func (e *Driver) NotExist(selector string) {
	e.Predicate(selector+" have to disappear", func() bool {
		return e.session.Count(selector) == 0
	})
}

// IsClosed check if tab was closed
func (e *Driver) IsClosed() bool {
	closed := e.session.IsClosed()
	e.log([]interface{}{}, []interface{}{closed})
	return closed
}

// Closed waiting for target closes
func (e *Driver) Closed() {
	e.Predicate("expect target closes", func() bool {
		return e.session.IsClosed()
	})
}

// Text get element's text (innerText or input's value)
func (e *Driver) Text(selector string) string {
	texts := e.TextAll(selector)
	if texts != nil && len(texts) > 0 {
		return texts[0]
	}
	return ""
}

// TextAll get elements text query by selector
func (e *Driver) TextAll(selector string) []string {
	var text []string
	var err error
	_ = retry(func() error {
		text, err = e.session.Text(selector)
		return err
	})
	e.log([]interface{}{selector}, []interface{}{text})
	if err != nil {
		e.Panic(err, "get text of elements `%s` failed", selector)
	}
	return text
}

// GetRectangle get dimension of element
func (e *Driver) GetRectangle(selector string) *cdp.Rect {
	var rect *cdp.Rect
	var err error
	_ = retry(func() error {
		rect, err = e.session.GetRectangle(selector)
		return err
	})
	e.log([]interface{}{selector}, []interface{}{rect})
	if err != nil {
		e.Panic(err, "get dimension of element `%s` failed", selector)
	}
	return rect
}

// Hover mouse hover on element
func (e *Driver) Hover(selector string) {
	err := retry(func() error {
		return e.session.Hover(selector)
	})
	e.log([]interface{}{selector}, []interface{}{})
	if err != nil {
		e.Panic(err, "hover on element `%s` failed", selector)
	}
}

// CallScriptOn evaluate javascript code on element
// executed as function() { `script` }, for access to element use `this`
// for example script = `this.innerText = "test"`
func (e *Driver) CallScriptOn(selector, script string) interface{} {
	var result interface{}
	var err error
	retry(func() error {
		result, err = e.session.CallScriptOn(selector, script)
		return err
	})
	e.log([]interface{}{selector, script}, []interface{}{result})
	if err != nil {
		e.Panic(err, "call script on element `%s` failed", selector)
	}
	return result
}

// SendKeys send keyboard keys to focused element
func (e *Driver) SendKeys(key ...rune) {
	err := e.session.SendKeys(key...)
	e.log([]interface{}{key}, []interface{}{})
	if err != nil {
		e.Panic(err, "send keys failed")
	}
}

// NewTarget call init and waiting for new target created
func (e *Driver) NewTarget(init func()) string {
	targetInfo := e.session.TargetCreated()
	e.log(nil, nil)
	init()
	select {
	case target := <-targetInfo:
		return target.TargetID
	case <-time.After(cdp.NavigationTimeout):
		e.Panic(nil, "new target has not been created")
		return ""
	}
}

// Count count of elements in DOM
func (e *Driver) Count(selector string) int {
	var count = e.session.Count(selector)
	e.log([]interface{}{selector}, []interface{}{count})
	return count
}
