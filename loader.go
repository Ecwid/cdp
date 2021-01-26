package cdp

import (
	"sync"
	"time"
)

type loader struct {
	*sync.Cond
	context int64
	frame   string
	mx      *sync.Mutex
	locked  bool
}

func newLoader() *loader {
	return &loader{
		context: 0,
		Cond:    sync.NewCond(&sync.Mutex{}),
		mx:      &sync.Mutex{},
	}
}

func (loader *loader) lock() {
	loader.L.Lock()
	loader.locked = true
	loader.L.Unlock()
}

func (loader *loader) unlock() {
	loader.L.Lock()
	loader.locked = false
	loader.L.Unlock()
	loader.Broadcast()
}

func condwait(cond *sync.Cond, deadline time.Duration) error {
	var c = make(chan struct{})
	go func() {
		cond.Wait()
		close(c)
	}()
	select {
	case <-c:
		return nil
	case <-time.After(deadline):
		return ErrTimeout
	}
}

func (loader *loader) wait(deadline time.Duration) (err error) {
	loader.L.Lock()
	for loader.locked {
		err = condwait(loader.Cond, deadline)
	}
	loader.L.Unlock()
	return err
}

func (session *Session) newContext(id string) (err error) {
	if err = session.loader.wait(session.deadline); err != nil {
		return err
	}
	session.loader.mx.Lock()
	defer session.loader.mx.Unlock()
	var c int64
	if id != session.target && id != "" {
		c, err = session.CreateIsolatedWorld(id, "my-current-frame-context")
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
