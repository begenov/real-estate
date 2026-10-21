package limiter

import (
	"context"
	"github.com/begenov/real-estate/internal/logger"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type rateLimiter struct {
	sync.RWMutex

	visitors map[string]*visitor
	limit    rate.Limit
	burst    int
	ttl      time.Duration
}

func newRateLimiter(rps, burst int, ttl time.Duration) *rateLimiter {
	return &rateLimiter{
		visitors: make(map[string]*visitor),
		limit:    rate.Limit(rps),
		burst:    burst,
		ttl:      ttl,
	}
}

func (l *rateLimiter) getVisitor(ip string) *rate.Limiter {
	l.Lock()
	v, exists := l.visitors[ip]
	if !exists {
		v = &visitor{rate.NewLimiter(l.limit, l.burst), time.Now()}
		l.visitors[ip] = v
	} else {
		v.lastSeen = time.Now()
	}
	l.Unlock()
	return v.limiter
}

func (l *rateLimiter) cleanupVisitors(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			l.Lock()
			for ip, v := range l.visitors {
				if time.Since(v.lastSeen) > l.ttl {
					delete(l.visitors, ip)
				}
			}
			l.Unlock()
		}
	}
}

func Limit(rps int, burst int, ttl time.Duration) gin.HandlerFunc {
	l := newRateLimiter(rps, burst, ttl)

	go l.cleanupVisitors(context.Background())

	return func(c *gin.Context) {
		ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		if err != nil {
			logger.Error("Failed to parse IP address:", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if !l.getVisitor(ip).Allow() {
			logger.Warn("Rate limit exceeded for IP:", ip)
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		c.Next()
	}
}
