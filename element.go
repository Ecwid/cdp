package cdp

import (
	"errors"
	"fmt"

	"github.com/ecwid/cdp/pkg/devtool"
)

// Element ...
type Element struct {
	session *Session
	ID      string
	context int64
}

func newElement(s *Session, parent *Element, ID string) (*Element, error) {
	c, err := s.currentContext()
	if err != nil {
		return nil, err
	}
	return &Element{
		ID:      ID,
		session: s,
		context: c,
	}, nil
}

func (e *Element) call(functionDeclaration string, arg ...interface{}) (*devtool.RemoteObject, error) {
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
	_, err := e.call(atomDispatchEvents, append([]string{}, events...))
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

// Click ...
func (e *Element) Click() error {
	if err := e.session.scrollIntoViewIfNeeded(e.ID); err != nil {
		return err
	}
	x, y, err := e.clickablePoint()
	if err != nil {
		return err
	}
	if _, err = e.call(atomPreventMissClick); err != nil {
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
	ok, err := e.call(atomClickDone)
	if err == nil {
		if ok.Bool() {
			return nil
		}
		return ErrMissClick
	}
	if err == ErrSessionAlreadyClosed || err == ErrStaleElementReference {
		return nil
	}
	return err
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
	if _, err := e.session.GetContentQuads(e.ID, false); err != nil {
		return false, err
	}
	val, err := e.call(atomIsVisible)
	if err != nil {
		return false, err
	}
	return val.Bool(), nil
}

// Hover hover mouse on element
func (e *Element) Hover() error {
	if err := e.session.scrollIntoViewIfNeeded(e.ID); err != nil {
		return err
	}
	x, y, err := e.clickablePoint()
	if err != nil {
		return err
	}
	return e.session.MouseMove(x, y)
}

// Type ...
func (e *Element) Type(text string) error {
	if err := e.session.scrollIntoViewIfNeeded(e.ID); err != nil {
		return err
	}
	if _, err := e.call(atomClearInput); err != nil {
		return err
	}
	if err := e.Focus(); err != nil {
		return err
	}
	for _, c := range text {
		if isKey(c) {
			if err := e.session.press(keyDefinitions[c]); err != nil {
				return err
			}
		} else {
			if err := e.session.InsertText(string(c)); err != nil {
				return err
			}
		}
	}
	return nil
}

// InsertText ...
func (e *Element) InsertText(text string, key ...rune) error {
	if err := e.session.scrollIntoViewIfNeeded(e.ID); err != nil {
		return err
	}
	v, err := e.IsVisible()
	if err != nil {
		return err
	}
	if !v {
		return ErrElementInvisible
	}
	if _, err = e.call(atomClearInput); err != nil {
		return err
	}
	if err = e.Focus(); err != nil {
		return err
	}
	// insert text, not typing
	if err = e.session.InsertText(text); err != nil {
		return err
	}
	if err := e.dispatchEvents("keypress", "input", "keyup", "change"); err != nil {
		return err
	}
	if key != nil {
		return e.session.SendKeys(key...)
	}
	return nil
}

// GetText ...
func (e *Element) GetText() (string, error) {
	v, err := e.call(atomGetInnerText)
	if err != nil {
		return "", err
	}
	return v.String()
}

// SetAttr ...
func (e *Element) SetAttr(attr string, value string) (err error) {
	_, err = e.call(atomSetAttr, attr, value)
	return
}

// GetAttr ...
func (e *Element) GetAttr(attr string) (string, error) {
	v, err := e.call(atomGetAttr, attr)
	if err != nil {
		return "", err
	}
	return v.String()
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
	v, err := e.call(atomGetComputedStyle, style)
	if err != nil {
		return "", err
	}
	return v.String()
}

// GetSelected ...
func (e *Element) GetSelected(selectedText bool) ([]string, error) {
	a := atomGetSelected
	if selectedText {
		a = atomGetSelectedText
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
		val, err := e.call(atomMutationObserver, attributes, childList, subtree)
		if err != nil {
			chanerr <- err
			return
		}
		str, err := val.String()
		if err != nil {
			chanerr <- err
		}
		mutation <- str
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
		return errors.New("specified element is not a SELECT")
	}
	contains, err := e.call(atomSelectContains, values)
	if err != nil {
		return err
	}
	if !contains.Bool() {
		return fmt.Errorf("select element has no options %s", values)
	}
	if _, err = e.call(atomSelect, values); err != nil {
		return err
	}
	if err := e.dispatchEvents("click", "input", "change"); err != nil {
		return err
	}
	return nil
}

// Checkbox Checkbox
func (e *Element) Checkbox(check bool) error {
	if _, err := e.call(atomCheckBox, check); err != nil {
		return err
	}
	if err := e.dispatchEvents("click", "input", "change"); err != nil {
		return err
	}
	return nil
}

// Checked ...
func (e *Element) Checked() (bool, error) {
	checked, err := e.call(atomChecked)
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
