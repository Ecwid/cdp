package cdp

import (
	"errors"
	"strings"
	"time"

	"github.com/ecwid/cdp/atom"
)

var (
	errElementNotDisplayed = errors.New("element not rendered or overlapped")
	errElementOverlapped   = errors.New("element overlapped")
	errClickNotConfirmed   = errors.New("click not confirmed")
	errTypeIsNotString     = errors.New("object type is not string")
)

func (session *Session) release(elements ...string) {
	for _, e := range elements {
		_ = session.releaseObject(e)
	}
}

func (session *Session) findElement(selector string) (string, error) {
	elements, err := session.findElements(selector)
	if err != nil {
		return "", err
	}
	return elements[0], nil
}

func (session *Session) findElements(selector string) ([]string, error) {
	selector = strings.ReplaceAll(selector, `"`, `\"`)
	array, err := session.Evaluate(`document.querySelectorAll("`+selector+`")`, session.contextID)
	if array == nil || array.Description == "NodeList(0)" {
		return nil, errors.New(`no one element ` + selector + ` found in frame ` + session.frameID)
	}
	descriptor, err := session.getProperties(array.ObjectID)
	if err != nil {
		return nil, err
	}
	var elements []string
	for _, d := range descriptor {
		if !d.Enumerable {
			continue
		}
		elements = append(elements, d.Value.ObjectID)
	}
	return elements, err
}

func (session *Session) dispatchEvent(element string, name string) error {
	_, err := session.callFunctionOn(element, atom.DispatchEvent, name)
	return err
}

// Upload upload files into selector
func (session *Session) Upload(selector string, files ...string) error {
	element, err := session.findElement(selector)
	if err != nil {
		return err
	}
	defer session.release(element)
	return session.setFileInputFiles(files, element)
}

func (session *Session) clickablePoint(objectID string) (x float64, y float64, e error) {
	rect, err := session.getContentQuads(0, objectID)
	if err != nil {
		return -1, -1, errElementNotDisplayed
	}
	_ = session.highlightQuad(rect, &rgba{R: 255, G: 1, B: 1})
	x, y = rect.middle()
	/*
		При определении элемента, по которому произойдет клик по координатам x, y, необходимо учесть случай,
		когда клик происходит во фрейме. Во фрейме метод elementFromPoint(x, y) работает относительно координат самого фрейма,
		а не абсолютных координат viewport браузера. Здесь мы вычисляем координаты фрейма и вычитаем их
		из значений x, y для теста клика.
	*/
	cX, cY := x, y
	if session.frameID != session.targetID {
		frameElement, err := session.getFrameOwner(session.frameID)
		if err != nil {
			return x, y, err
		}
		size, err := session.getContentQuads(frameElement, "")
		if err != nil {
			return x, y, err
		}
		cX -= size[0]
		cY -= size[1]
	}
	// Выполняет тест клика по координатам cX, cY (координаты относительно текущего фрейма)
	// Если клик принимает другой элемент, либо элемент не родитель ожидаемого, то выбросим ошибк
	clickable, err := session.callFunctionOn(objectID, atom.IsClickableAt, cX, cY)
	if err != nil || !clickable.bool() {
		return x, y, errElementOverlapped
	}
	return x, y, nil
}

// Click ...
func (session *Session) Click(selector string) error {
	// find element for click at
	element, err := session.findElement(selector)
	if err != nil {
		return err
	}
	// release javascript object on exit
	defer session.release(element)
	// scroll element into view
	if _, err = session.callFunctionOn(element, atom.ScrollIntoView); err != nil {
		return err
	}
	// add click event listener on element
	_, err = session.callFunctionOn(element, atom.AddEventFired, "click")
	if err != nil {
		return err
	}
	session.dispatchMouseEvent(-1, -1, DispatchMouseEventMoved, "none")
	// calculate click point
	x, y, err := session.clickablePoint(element)
	if err != nil {
		return err
	}

	session.dispatchMouseEvent(x, y, DispatchMouseEventMoved, "none")
	session.dispatchMouseEvent(x, y, DispatchMouseEventPressed, "left")
	session.dispatchMouseEvent(x, y, DispatchMouseEventReleased, "left")

	// check to click happens
	fired, err := session.callFunctionOn(element, atom.IsEventFired)
	if err != nil || fired.bool() {
		return nil
	}
	return errClickNotConfirmed
}

// HoverXY move mouse at (x, y)
func (session *Session) HoverXY(x, y float64) {
	session.dispatchMouseEvent(x, y, DispatchMouseEventMoved, "none")
}

// Hover hover mouse on element
func (session *Session) Hover(selector string) error {
	objectID, err := session.findElement(selector)
	if err != nil {
		return err
	}
	defer session.release(objectID)
	if _, err = session.callFunctionOn(objectID, atom.ScrollIntoView); err != nil {
		return err
	}
	session.dispatchMouseEvent(0, 0, DispatchMouseEventMoved, "none")
	q, err := session.getContentQuads(0, objectID)
	if err != nil {
		return errElementNotDisplayed
	}
	session.HoverXY(q.middle())
	return nil
}

// Type ...
func (session *Session) Type(selector string, text string, key ...rune) error {
	// find element for type in
	element, err := session.findElement(selector)
	if err != nil {
		return err
	}
	// release javascript object on exit
	defer session.release(element)
	// element needs focus to send keys
	if err = session.focus(element); err != nil {
		return err
	}
	// clear input before type text
	if _, err := session.callFunctionOn(element, atom.ClearInput); err != nil {
		return err
	}
	// insert text, not typing
	err = session.insertText(text)
	if err != nil {
		return err
	}
	session.dispatchEvent(element, `input`)
	session.dispatchEvent(element, `change`)

	// send keyboard key after some pause
	if key != nil {
		time.Sleep(time.Millisecond * 200)
		for _, k := range key {
			session.sendRune(k)
		}
	}
	return nil
}

func (session *Session) textContent(objectID string) (string, error) {
	obj, err := session.callFunctionOn(objectID, atom.GetInnerText)
	if err != nil {
		return "", err
	}
	if obj.Type != "string" {
		return "", errTypeIsNotString
	}
	return obj.Value.(string), nil
}

// Text ...
func (session *Session) Text(selector string) ([]string, error) {
	elements, err := session.findElements(selector)
	if err != nil {
		return nil, err
	}
	defer session.release(elements...)
	array := make([]string, len(elements))
	for index, el := range elements {
		array[index], err = session.textContent(el)
		if err != nil {
			return nil, err
		}
	}
	return array, nil
}

// SetAttr ...
func (session *Session) SetAttr(selector string, attr string, value string) error {
	element, err := session.findElement(selector)
	if err != nil {
		return err
	}
	defer session.release(element)
	_, err = session.callFunctionOn(element, atom.SetAttr, attr, value)
	if err != nil {
		return err
	}
	return nil
}

// GetAttr ...
func (session *Session) GetAttr(selector string, attr string) (string, error) {
	element, err := session.findElement(selector)
	if err != nil {
		return "", err
	}
	defer session.release(element)
	value, err := session.callFunctionOn(element, atom.GetAttr, attr)
	if err != nil {
		return "", err
	}
	if value.Type != "string" {
		return "", errTypeIsNotString
	}
	return value.Value.(string), nil
}

// GetRectangle ...
func (session *Session) GetRectangle(selector string) (*Rect, error) {
	element, err := session.findElement(selector)
	if err != nil {
		return nil, err
	}
	defer session.release(element)
	q, err := session.getContentQuads(0, element)
	if err != nil {
		if err.Error() == `Could not compute content quads.` {
			// element not visible and dimension can't be got
			return nil, nil
		}
		return nil, err
	}
	rect := &Rect{
		X:      q[0],
		Y:      q[1],
		Width:  q[2] - q[0],
		Height: q[7] - q[1],
	}
	return rect, nil
}

// GetComputedStyle ...
func (session *Session) GetComputedStyle(selector string, style string) (string, error) {
	element, err := session.findElement(selector)
	if err != nil {
		return "", err
	}
	defer session.release(element)
	computedStyle, err := session.callFunctionOn(element, atom.GetComputedStyle, style)
	if err != nil {
		return "", err
	}
	if computedStyle.Type != "string" {
		return "", errTypeIsNotString
	}
	return computedStyle.Value.(string), nil
}

// GetSelected ...
func (session *Session) GetSelected(selector string, selectedText bool) ([]string, error) {
	element, err := session.findElement(selector)
	if err != nil {
		return nil, err
	}
	defer session.release(element)
	a := atom.GetSelected
	if selectedText {
		a = atom.GetSelectedText
	}
	selected, err := session.callFunctionOn(element, a)
	descriptor, err := session.getProperties(selected.ObjectID)
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

// Select ...
func (session *Session) Select(selector string, values ...string) error {
	element, err := session.findElement(selector)
	if err != nil {
		return err
	}
	defer session.release(element)
	node, err := session.describeNode(element)
	if err != nil {
		return err
	}
	if "SELECT" != node.NodeName {
		return errors.New(`node ` + selector + ` require has type SELECT`)
	}
	has, err := session.callFunctionOn(element, atom.SelectHasOptions, values)
	if !has.bool() {
		return errors.New(`select ` + selector + ` doesn't has options with values ` + strings.Join(values, ","))
	}
	_, err = session.callFunctionOn(element, atom.Select, values)
	return err
}

// Checkbox Checkbox
func (session *Session) Checkbox(selector string, check bool) error {
	element, err := session.findElement(selector)
	if err != nil {
		return err
	}
	defer session.release(element)
	if _, err = session.callFunctionOn(element, atom.CheckBox, check); err != nil {
		return err
	}
	time.Sleep(time.Millisecond * 250) // todo
	session.dispatchEvent(element, `click`)
	session.dispatchEvent(element, `change`)
	return nil
}

// IsChecked ...
func (session *Session) IsChecked(selector string) (bool, error) {
	element, err := session.findElement(selector)
	if err != nil {
		return false, err
	}
	defer session.release(element)
	checked, err := session.callFunctionOn(element, atom.IsChecked)
	return checked.bool(), err
}

// Count count of elements
func (session *Session) Count(selector string) int {
	elements, err := session.findElements(selector)
	if err != nil {
		return 0
	}
	defer session.release(elements...)
	return len(elements)
}
