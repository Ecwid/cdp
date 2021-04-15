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
	return fmt.Sprintf("no such element %s", e.selector)
}

// cdp errors
var (
	ErrStaleElementReference  = errors.New("referenced element is no longer attached to the DOM") // cannot find context with specified id
	ErrNoPageTarget           = errors.New("no target with page type found")
	ErrElementInvisible       = errors.New("element invisible")
	ErrElementIsOutOfViewport = errors.New("element is out of viewport")
	ErrMissClick              = errors.New("click is not received by element")
	ErrObjectNotString        = errors.New("object type is not string")
	ErrSessionAlreadyClosed   = errors.New("session was already closed")
	ErrConnectionClosed       = errors.New("websocket connection abnormal closure, browser may have died")
	ErrTargetCreatedTimeout   = errors.New("target creation timeout was reached")
	ErrLoadTimeout            = errors.New("load state timeout was reached")
	ErrContextDetached        = errors.New("frame was detached")
)
