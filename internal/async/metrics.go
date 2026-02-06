package async

import (
	"context"
	"github.com/begenov/real-estate/internal/logger"
	"time"
)

func WithTiming(name string, fn func(context.Context) error) func(context.Context) error {
	return func(ctx context.Context) error {
		start := time.Now()
		err := fn(ctx)
		logger.Infof("goroutine %s finished in %s", name, time.Since(start))
		return err
	}
}
