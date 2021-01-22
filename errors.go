package cdp

import (
	"errors"
	"fmt"
)

// NoSuchElementError ..
type NoSuchElementError struct {
	selector string
	context  int64
	frame    string
}

func (e NoSuchElementError) Error() string {
	return fmt.Sprintf("no such element %s\ncontext: %d, frame %s", e.selector, e.context, e.frame)
}

// IsElementError ...
func IsElementError(err error) bool {
	return false
}

// cdp errors
var (
	ErrElementDetached = errors.New("referenced element is no longer attached to the DOM")
	// ErrNoSuchContext   = errors.New("no context with given id found")
	// ErrNoSuchElement          = errors.New("no such element")
	ErrNoSuchFrame            = errors.New("no such frame")
	ErrFrameDetached          = errors.New("frame you working on was detached")
	ErrNoPageTarget           = errors.New("no one page target")
	ErrDevtoolTimeout         = errors.New("devtool response reached timeout")
	ErrNavigateTimeout        = errors.New("navigation reached timeout")
	ErrElementInvisible       = errors.New("element invisible")
	ErrElementIsOutOfViewport = errors.New("element is out of viewport")
	ErrElementNotFocusable    = errors.New("element is not focusable")
	ErrElementMissClick       = errors.New("miss click - element is overlapping or changing its coordinates")
	ErrInvalidString          = errors.New("object type is not string")
	ErrInvalidElementFrame    = errors.New("specified element is not a IFRAME")
	ErrInvalidElementSelect   = errors.New("specified element is not a SELECT")
	ErrInvalidElementOption   = errors.New("specified element has no options")
	ErrSessionClosed          = errors.New("session closed")
	ErrWebSocketClosure       = errors.New("abnormal closure")
	ErrTimeout                = errors.New("timeout")
)
