package cdp

import "time"

// Input events
const (
	dispatchKeyEventChar       = "char"
	dispatchKeyEventKeyDown    = "keyDown"
	dispatchKeyEventKeyUp      = "keyUp"
	dispatchMouseEventMoved    = "mouseMoved"
	dispatchMouseEventPressed  = "mousePressed"
	dispatchMouseEventReleased = "mouseReleased"
)

// MouseMove ...
func (i Input) MouseMove(x, y float64) error {
	return i.dispatchMouseEvent(x, y, dispatchMouseEventMoved, "none")
}

func (i Input) sendRune(c rune) error {
	if err := i.call("Input.dispatchKeyEvent", Map{
		"type":                  dispatchKeyEventKeyDown,
		"windowsVirtualKeyCode": int(c),
		"nativeVirtualKeyCode":  int(c),
		"unmodifiedText":        string(c),
		"text":                  string(c),
	}, nil); err != nil {
		return err
	}
	return i.call("Input.dispatchKeyEvent", Map{
		"type":                  dispatchKeyEventKeyUp,
		"windowsVirtualKeyCode": int(c),
		"nativeVirtualKeyCode":  int(c),
		"unmodifiedText":        string(c),
		"text":                  string(c),
	}, nil)
}

func (i Input) dispatchKeyEvent(text string) error {
	for _, c := range text {
		time.Sleep(time.Millisecond * 10)
		err := i.call("Input.dispatchKeyEvent", Map{
			"type":                  dispatchKeyEventChar,
			"windowsVirtualKeyCode": int(c),
			"nativeVirtualKeyCode":  int(c),
			"unmodifiedText":        string(c),
			"text":                  string(c),
		}, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// InsertText method emulates inserting text that doesn't come from a key press, for example an emoji keyboard or an IME
func (i Input) InsertText(text string) error {
	return i.call("Input.insertText", Map{"text": text}, nil)
}

func (i Input) dispatchMouseEvent(x float64, y float64, eventType string, button string) error {
	return i.call("Input.dispatchMouseEvent", Map{
		"type":       eventType,
		"button":     button,
		"x":          x,
		"y":          y,
		"clickCount": 1,
	}, nil)
}

// SendKeys send keyboard keys to focused element
func (i Input) SendKeys(key ...rune) error {
	var err error
	for _, k := range key {
		err = i.sendRune(k)
		if err != nil {
			return err
		}
	}
	return nil
}
