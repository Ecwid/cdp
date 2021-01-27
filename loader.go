package cdp

import (
	"sync"
	"time"
)

type loader struct {
	cond    *sync.Cond
	context int64
	frame   string
	mx      *sync.Mutex
	locked  bool
}

func newLoader() *loader {
	return &loader{
		context: 0,
		cond:    sync.NewCond(&sync.Mutex{}),
		mx:      &sync.Mutex{},
	}
}

func (l *loader) lock() {
	l.cond.L.Lock()
	l.locked = true
	l.cond.L.Unlock()
}

func (l *loader) unlock() {
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
			err = ErrTimeout
			cond.Broadcast()
		}
	}
}

func (l *loader) wait(deadline time.Duration) (err error) {
	l.cond.L.Lock()
	for l.locked {
		err = condwait(l.cond, deadline)
	}
	l.cond.L.Unlock()
	return err
}

func (session *Session) newContext(frameID string) (err error) {
	if err = session.loader.wait(session.deadline); err != nil {
		return err
	}
	session.loader.mx.Lock()
	defer session.loader.mx.Unlock()
	var c int64
	if frameID != session.target && frameID != "" {
		c, err = session.CreateIsolatedWorld(frameID, "my-current-frame-context")
		if err != nil {
			return err
		}
	}
	session.loader.context = c
	return nil
}

func (session *Session) currentContext() int64 {
	session.loader.mx.Lock()
	defer session.loader.mx.Unlock()
	return session.loader.context
}
