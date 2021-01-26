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
	return fmt.Sprintf("no such element `%s` context: %d, frame %s", e.selector, e.context, e.frame)
}

// cdp errors
var (
	ErrStaleElementReference = errors.New("referenced element is no longer attached to the DOM") // cannot find context with specified id
	// ErrElementDetached        = errors.New("referenced element is no longer attached to the DOM")
	// ErrNoSuchFrame            = errors.New("no such frame")
	// ErrFrameDetached          = errors.New("frame you working on was detached")
	// ErrNoPageTarget           = errors.New("no one page target")
	ErrElementInvisible       = errors.New("element invisible")
	ErrElementIsOutOfViewport = errors.New("element is out of viewport")
	ErrClickFailed            = errors.New("element is overlapping or change position")
	ErrInvalidString          = errors.New("object type is not string")
	ErrInvalidElementSelect   = errors.New("specified element is not a SELECT")
	ErrInvalidElementOption   = errors.New("specified element has no options")
	ErrSessionClosed          = errors.New("session closed")
	ErrConnectionClosed       = errors.New("abnormal closure")
	ErrTimeout                = errors.New("timeout")
	// ErrContextDestroyed       = errors.New("cannot find context with specified id")
)
