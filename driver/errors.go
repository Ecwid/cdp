package driver

import "fmt"

// ErrDriver driver error
type ErrDriver struct {
	SessionID string
	Message   string
	Cause     error
}

func (e ErrDriver) Error() string {
	return fmt.Sprintf("sessionID: %s\n error: %s\n\tcause: %s", e.SessionID, e.Message, e.Cause.Error())
}
