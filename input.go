package cdp

import "time"

// Input events
const (
	DispatchKeyEventChar       = "char"
	DispatchKeyEventKeyDown    = "keyDown"
	DispatchKeyEventKeyUp      = "keyUp"
	DispatchMouseEventMoved    = "mouseMoved"
	DispatchMouseEventPressed  = "mousePressed"
	DispatchMouseEventReleased = "mouseReleased"
)

func (session *Session) sendRune(c rune) {
	_, _ = session.blockingSend("Input.dispatchKeyEvent", &Params{
		"type":                  DispatchKeyEventKeyDown,
		"windowsVirtualKeyCode": int(c),
		"nativeVirtualKeyCode":  int(c),
		"unmodifiedText":        string(c),
		"text":                  string(c),
	})
	_, _ = session.blockingSend("Input.dispatchKeyEvent", &Params{
		"type":                  DispatchKeyEventKeyUp,
		"windowsVirtualKeyCode": int(c),
		"nativeVirtualKeyCode":  int(c),
		"unmodifiedText":        string(c),
		"text":                  string(c),
	})
}

func (session *Session) dispatchKeyEvent(text string) error {
	for _, c := range text {
		time.Sleep(time.Millisecond * 10)
		if _, err := session.blockingSend("Input.dispatchKeyEvent", &Params{
			"type":                  DispatchKeyEventChar,
			"windowsVirtualKeyCode": int(c),
			"nativeVirtualKeyCode":  int(c),
			"unmodifiedText":        string(c),
			"text":                  string(c),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) insertText(text string) error {
	_, err := session.blockingSend("Input.insertText", &Params{"text": text})
	return err
}
func (session *Session) dispatchMouseEvent(x float64, y float64, eventType string, button string) {
	_, _ = session.blockingSend("Input.dispatchMouseEvent", &Params{
		"type":       eventType,
		"button":     button,
		"x":          x,
		"y":          y,
		"clickCount": 1,
	})
}
