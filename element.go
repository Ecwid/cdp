package cdp

import (
	"errors"
	"time"

	"github.com/ecwid/cdp/internal/atom"
	"github.com/ecwid/cdp/pkg/devtool"
)

// Element ...
type Element struct {
	session *Session
	ID      string
	context int64
}

func newElement(s *Session, parent *Element, ID string) (*Element, error) {
	c, err := s.executionContext()
	if err != nil {
		return nil, err
	}
	el := &Element{
		ID:      ID,
		session: s,
		context: c,
	}
	return el, nil
}

// Detached ...
func (e *Element) Detached() bool {
	c, err := e.session.executionContext()
	if err != nil {
		return true
	}
	return e.context != c
}

func (e *Element) call(functionDeclaration string, arg ...interface{}) (*devtool.RemoteObject, error) {
	if e.Detached() {
		return nil, ErrElementDetached
	}
	return e.session.callFunctionOn(e.ID, functionDeclaration, arg...)
}

// Call evaluate javascript for element, for example `function() {return this.innerHTML}`
func (e *Element) Call(functionDeclaration string, arg ...interface{}) (interface{}, error) {
	v, err := e.call(functionDeclaration, arg...)
	if err != nil {
		return nil, err
	}
	return v.Value, nil
}

func (e *Element) dispatchEvents(events ...string) error {
	_, err := e.call(atom.DispatchEvents, append([]string{}, events...))
	return err
}

// Focus focus element
func (e *Element) Focus() error {
	return e.session.call("DOM.focus", Map{"objectId": e.ID}, nil)
}

// Upload upload files
func (e *Element) Upload(files ...string) error {
	return e.session.call("DOM.setFileInputFiles", Map{"files": files, "objectId": e.ID}, nil)
}

func (e *Element) clickablePoint() (x float64, y float64, err error) {
	r, err := e.session.GetContentQuads(e.ID, true)
	if err != nil {
		return -1, -1, err
	}
	x, y = r.Middle()
	return x, y, nil
}

func (e *Element) scrollIntoView() (err error) {
	_, err = e.call(atom.ScrollIntoView)
	return
}

// Click ...
func (e *Element) Click() error {
	if err := e.scrollIntoView(); err != nil {
		return err
	}
	x, y, err := e.clickablePoint()
	if err != nil {
		return err
	}
	if _, err = e.call(atom.PreventMissClick); err != nil {
		return err
	}
	if err = e.session.dispatchMouseEvent(x, y, dispatchMouseEventMoved, "none"); err != nil {
		return err
	}
	if err = e.session.dispatchMouseEvent(x, y, dispatchMouseEventPressed, "left"); err != nil {
		return err
	}
	if err = e.session.dispatchMouseEvent(x, y, dispatchMouseEventReleased, "left"); err != nil {
		return err
	}
	hit, err := e.call(atom.ClickHitReturn)
	// in case when click is initiate navigation which destroyed context of element (ErrElementDetached)
	// or click may closes a popup (ErrSessionClosed)
	switch err {
	case ErrElementDetached, ErrSessionClosed:
		return nil
	case nil:
		if hit.Bool() {
			return nil
		}
		return ErrClickFailed
	default:
		return err
	}
}

// GetFrameID get if for IFRAME element
func (e *Element) GetFrameID() (string, error) {
	node, err := e.session.GetNode(e.ID)
	if err != nil {
		return "", err
	}
	if "IFRAME" != node.NodeName && "FRAME" != node.NodeName {
		return "", errors.New("specified element is not a IFRAME")
	}
	return node.FrameID, nil
}

// IsVisible is element visible (element has area that clickable in viewport)
func (e *Element) IsVisible() (bool, error) {
	if _, _, err := e.clickablePoint(); err != nil {
		if err == ErrElementInvisible {
			return false, nil
		}
		return false, err
	}
	if vis, err := e.call(atom.IsVisible); err != nil || !vis.Bool() {
		return false, nil
	}
	return true, nil
}

// Hover hover mouse on element
func (e *Element) Hover() error {
	if err := e.scrollIntoView(); err != nil {
		return err
	}
	x, y, err := e.clickablePoint()
	if err != nil {
		return err
	}
	return e.session.MouseMove(x, y)
}

// Clear ...
func (e *Element) Clear() error {
	var err error
	if err = e.Focus(); err != nil {
		return err
	}
	_, err = e.call(atom.ClearInput)
	return err
}

// Type ...
func (e *Element) Type(text string, key ...rune) error {
	var err error
	if enable, err := e.call(atom.IsVisible); err != nil || !enable.Bool() {
		return err
	}
	if err = e.Clear(); err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 250)
	if err := e.dispatchEvents("keydown"); err != nil {
		return err
	}
	// insert text, not typing
	// todo natural typing
	err = e.session.InsertText(text)
	if err != nil {
		return err
	}
	if err := e.dispatchEvents("keypress", "input", "keyup", "change"); err != nil {
		return err
	}
	// send keyboard key after some pause
	if key != nil {
		time.Sleep(time.Millisecond * 250)
		return e.session.SendKeys(key...)
	}
	return nil
}

func (e *Element) string(functionDeclaration string, arg ...interface{}) (string, error) {
	res, err := e.call(functionDeclaration, arg...)
	if err != nil {
		return "", err
	}
	if res.Type != "string" {
		return "", ErrInvalidString
	}
	return res.Value.(string), nil
}

// GetText ...
func (e *Element) GetText() (string, error) {
	return e.string(atom.GetInnerText)
}

// SetAttr ...
func (e *Element) SetAttr(attr string, value string) (err error) {
	_, err = e.call(atom.SetAttr, attr, value)
	return
}

// GetAttr ...
func (e *Element) GetAttr(attr string) (string, error) {
	return e.string(atom.GetAttr, attr)
}

// GetRectangle ...
func (e *Element) GetRectangle() (*devtool.Rect, error) {
	q, err := e.session.GetContentQuads(e.ID, false)
	if err != nil {
		return nil, err
	}
	rect := &devtool.Rect{
		X:      q[0].X,
		Y:      q[0].Y,
		Width:  q[1].X - q[0].X,
		Height: q[3].Y - q[0].Y,
	}
	return rect, nil
}

// GetComputedStyle ...
func (e *Element) GetComputedStyle(style string) (string, error) {
	return e.string(atom.GetComputedStyle, style)
}

// GetSelected ...
func (e *Element) GetSelected(selectedText bool) ([]string, error) {
	a := atom.GetSelected
	if selectedText {
		a = atom.GetSelectedText
	}
	ro, err := e.call(a)
	if err != nil {
		return nil, err
	}
	descriptor, err := e.session.getProperties(ro.ObjectID)
	if err != nil {
		return nil, err
	}
	var options []string
	for _, d := range descriptor {
		if !d.Enumerable {
			continue
		}
		options = append(options, d.Value.Value.(string))
	}
	return options, nil
}

// ObserveMutation create MutationObserver Promise for element, returns type of first mutation
func (e *Element) ObserveMutation(attributes, childList, subtree bool) (chan string, chan error) {
	chanerr := make(chan error, 1)
	mutation := make(chan string, 1)
	go func() {
		val, err := e.call(atom.MutationObserver, attributes, childList, subtree)
		if err != nil {
			chanerr <- err
			return
		}
		if val.Type != "string" {
			chanerr <- ErrInvalidString
		}
		mutation <- val.Value.(string)
	}()
	return mutation, chanerr
}

// Select ...
func (e *Element) Select(values ...string) error {
	node, err := e.session.GetNode(e.ID)
	if err != nil {
		return err
	}
	if "SELECT" != node.NodeName {
		return ErrInvalidElementSelect
	}
	has, err := e.call(atom.SelectHasOptions, values)
	if err != nil {
		return err
	}
	if !has.Bool() {
		return ErrInvalidElementOption
	}
	if err = e.scrollIntoView(); err != nil {
		return err
	}
	if _, err = e.call(atom.Select, values); err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 250)
	if err := e.dispatchEvents("input", "change"); err != nil {
		return err
	}
	return nil
}

// Checkbox Checkbox
func (e *Element) Checkbox(check bool) error {
	if _, err := e.call(atom.CheckBox, check); err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 250)
	if err := e.dispatchEvents("click", "change"); err != nil {
		return err
	}
	return nil
}

// Checked ...
func (e *Element) Checked() (bool, error) {
	checked, err := e.call(atom.Checked)
	return checked.Bool(), err
}

// GetEventListeners returns event listeners of the given object.
func (e *Element) GetEventListeners() ([]string, error) {
	events := new(devtool.EventListeners)
	err := e.session.call("DOMDebugger.getEventListeners", Map{
		"objectId": e.ID,
		"depth":    1,
		"pierce":   true,
	}, events)
	if err != nil {
		return nil, err
	}
	types := make([]string, len(events.Listeners))
	for n, e := range events.Listeners {
		types[n] = e.Type
	}
	return types, nil
}

// Query ...
func (e *Element) Query(selector string) (*Element, error) {
	return e.session.query(e, selector)
}

// QueryAll ...
func (e *Element) QueryAll(selector string) ([]*Element, error) {
	return e.session.queryAll(e, selector)
}
