package driver

import "time"

type backoff struct {
	Max   time.Duration
	Delay time.Duration
	Retry func() error
}

// Do ...
func (b *backoff) Do() error {
	var err error
	for start := time.Now(); time.Since(start) < b.Max; {
		err = b.Retry()
		if err == nil {
			return nil
		}
		time.Sleep(b.Delay)
	}
	return err
}
