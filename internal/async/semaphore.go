package async

import "context"

type Semaphore struct {
	ch chan struct{}
}

func NewSemaphore(limit int) *Semaphore {
	if limit < 1 {
		limit = 1
	}
	return &Semaphore{ch: make(chan struct{}, limit)}
}

func (s *Semaphore) Acquire(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Semaphore) TryAcquire() bool {
	select {
	case s.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

func (s *Semaphore) Release() {
	select {
	case <-s.ch:
	default:
	}
}
