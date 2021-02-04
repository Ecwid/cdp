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
func (input Input) MouseMove(x, y float64) error {
	return input.dispatchMouseEvent(x, y, dispatchMouseEventMoved, "none")
}

func (input Input) press(k keyDefinition) error {
	if k.text == "" {
		k.text = k.key
	}
	p := Map{
		"type":                  dispatchKeyEventKeyDown,
		"windowsVirtualKeyCode": k.keyCode,
		"text":                  k.text,
	}
	if err := input.call("Input.dispatchKeyEvent", p, nil); err != nil {
		return err
	}
	p = Map{
		"type": dispatchKeyEventKeyUp,
		"text": k.text,
	}
	return input.call("Input.dispatchKeyEvent", p, nil)
}

// InsertText method emulates inserting text that doesn't come from a key press, for example an emoji keyboard or an IME
func (input Input) InsertText(text string) error {
	return input.call("Input.insertText", Map{"text": text}, nil)
}

func (input Input) dispatchMouseEvent(x float64, y float64, eventType string, button string) error {
	return input.call("Input.dispatchMouseEvent", Map{
		"type":       eventType,
		"button":     button,
		"x":          x,
		"y":          y,
		"clickCount": 1,
	}, nil)
}
