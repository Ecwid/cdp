package cdp

// Input events
const (
	dispatchKeyEventChar       = "char"
	dispatchKeyEventKeyDown    = "keyDown"
	dispatchKeyEventRawKeyDown = "rawKeyDown"
	dispatchKeyEventKeyUp      = "keyUp"
	dispatchMouseEventMoved    = "mouseMoved"
	dispatchMouseEventPressed  = "mousePressed"
	dispatchMouseEventReleased = "mouseReleased"
)

// MouseMove ...
func (session Input) MouseMove(x, y float64) error {
	return session.dispatchMouseEvent(x, y, dispatchMouseEventMoved, "none")
}

// Press ...
func (session Input) Press(c rune) error {
	return session.press(keyDefinition{keyCode: int(c), text: string(c)})
}

func (session Input) press(k keyDefinition) error {
	if k.text == "" {
		k.text = k.key
	}
	p := Map{
		"type":                  dispatchKeyEventKeyDown,
		"key":                   k.key,
		"code":                  k.code,
		"windowsVirtualKeyCode": k.keyCode,
		"text":                  k.text,
	}
	if err := session.call("Input.dispatchKeyEvent", p, nil); err != nil {
		return err
	}
	p = Map{
		"key":  k.key,
		"code": k.code,
		"type": dispatchKeyEventKeyUp,
		"text": k.text,
	}
	return session.call("Input.dispatchKeyEvent", p, nil)
}

// InsertText method emulates inserting text that doesn't come from a key press, for example an emoji keyboard or an IME
func (session Input) InsertText(text string) error {
	return session.call("Input.insertText", Map{"text": text}, nil)
}

func (session Input) dispatchMouseEvent(x float64, y float64, eventType string, button string) error {
	return session.call("Input.dispatchMouseEvent", Map{
		"type":       eventType,
		"button":     button,
		"x":          x,
		"y":          y,
		"clickCount": 1,
	}, nil)
}
