package async

import (
	"context"
	"errors"
	"sync"
)

var ErrPoolClosed = errors.New("worker pool is closed")

type WorkerPool struct {
	jobs      chan func(context.Context) error
	wg        sync.WaitGroup
	closeOnce sync.Once

	mu     sync.Mutex
	closed bool
}

func NewWorkerPool(concurrency, queue int) *WorkerPool {
	if concurrency < 1 {
		concurrency = 1
	}
	if queue < 0 {
		queue = 0
	}
	return &WorkerPool{
		jobs: make(chan func(context.Context) error, queue),
	}
}

func (p *WorkerPool) Start(ctx context.Context, concurrency int) {
	if concurrency < 1 {
		concurrency = 1
	}
	for i := 0; i < concurrency; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case job, ok := <-p.jobs:
					if !ok {
						return
					}
					_ = job(ctx)
				}
			}
		}()
	}
}

func (p *WorkerPool) Submit(ctx context.Context, job func(context.Context) error) error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return ErrPoolClosed
	}
	p.mu.Unlock()

	select {
	case p.jobs <- job:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *WorkerPool) Close() {
	p.closeOnce.Do(func() {
		p.mu.Lock()
		p.closed = true
		close(p.jobs)
		p.mu.Unlock()
	})
}

func (p *WorkerPool) Wait() {
	p.wg.Wait()
}
