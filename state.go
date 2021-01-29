package cdp

import (
	"sync"
	"time"
)

type state struct {
	cond    *sync.Cond
	locked  bool
	mx      *sync.Mutex
	context int64
}

func newState() *state {
	return &state{
		context: 0,
		cond:    sync.NewCond(&sync.Mutex{}),
		mx:      &sync.Mutex{},
	}
}

func (l *state) lock() {
	l.cond.L.Lock()
	l.locked = true
	l.cond.L.Unlock()
}

func (l *state) unlock() {
	l.cond.L.Lock()
	l.locked = false
	l.cond.Broadcast()
	l.cond.L.Unlock()
}

func condwait(cond *sync.Cond, deadline time.Duration) (err error) {
	var c = make(chan struct{})
	go func() {
		cond.Wait() // unlock of unlocked mutex
		close(c)
	}()
	for {
		select {
		case <-c:
			return err
		case <-time.After(deadline):
			err = ErrLoadTimeout
			cond.Broadcast()
		}
	}
}

func (l *state) wait(deadline time.Duration) error {
	l.cond.L.Lock()
	defer l.cond.L.Unlock()
	for l.locked {
		if err := condwait(l.cond, deadline); err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) newContext(frameID string) (err error) {
	if err = session.state.wait(session.deadline); err != nil {
		return err
	}
	session.state.mx.Lock()
	defer session.state.mx.Unlock()
	var c int64
	if frameID != session.target && frameID != "" {
		c, err = session.CreateIsolatedWorld(frameID, "my-current-frame-context")
		if err != nil {
			return err
		}
	}
	session.state.context = c
	return nil
}

func (session *Session) currentContext() int64 {
	session.state.mx.Lock()
	defer session.state.mx.Unlock()
	return session.state.context
}
