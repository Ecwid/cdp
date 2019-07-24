package driver

import "fmt"

// ErrDriver driver error
type ErrDriver struct {
	SessionID string
	Message   string
	Cause     error
}

func (e ErrDriver) Error() string {
	cause := ""
	if e.Cause != nil {
		cause = fmt.Sprintf("\n\tcause: %s", e.Cause.Error())
	}
	return fmt.Sprintf("sessionID: %s\nerror: %s%s", e.SessionID, e.Message, cause)
}
