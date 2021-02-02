package cdp

import (
	"sync"
)

type state struct {
	sync.Mutex
	context int64
	frame   string
}

func newState() *state {
	return &state{
		context: 0,
		frame:   "",
		Mutex:   sync.Mutex{},
	}
}

func (l *state) GetFrame() string {
	l.Lock()
	defer l.Unlock()
	f := l.frame
	return f
}

func (l *state) set(frameID string, contextID int64) {
	l.Lock()
	defer l.Unlock()
	l.frame = frameID
	l.context = contextID
}

func (l *state) reset() {
	l.set("", 0)
}

func (l *state) destroy() {
	l.set(l.GetFrame(), -1)
}

func (session *Session) currentContext() (int64, error) {
	session.state.Lock()
	defer session.state.Unlock()
	if session.state.context == -1 {
		var err error
		session.state.context, err = session.createContext(session.state.frame)
		if err != nil {
			return -1, err
		}
	}
	return session.state.context, nil
}

func (session *Session) createContext(frameID string) (int64, error) {
	if frameID != "" {
		session.ws.printf(LevelSessionState, "create_context for %s", frameID)
		return session.createIsolatedWorld(frameID, "my-current-frame-context")
	}
	return 0, nil
}
