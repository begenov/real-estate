package async

import (
	"context"
	"sync"
)

type Group struct {
	ctx    context.Context
	cancel context.CancelFunc

	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
}

func WithContext(ctx context.Context) (*Group, context.Context) {
	groupCtx, cancel := context.WithCancel(ctx)
	return &Group{
		ctx:    groupCtx,
		cancel: cancel,
	}, groupCtx
}

func (g *Group) Go(fn func(context.Context) error) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		if err := fn(g.ctx); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				g.cancel()
			})
		}
	}()
}

func (g *Group) Wait() error {
	g.wg.Wait()
	return g.err
}
