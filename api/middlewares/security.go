package middlewares

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type RequestLimiterConfig struct {
	RequestLimit int
	Interval     time.Duration
	BlockTime    time.Duration
}

var ipRateLimiter struct {
	sync.Mutex
	RequestCounts map[string]*RequestCounter
	Blacklist     map[string]time.Time
}

type RequestCounter struct {
	Count int
	Since time.Time
}

func InitIPRateLimiter() {
	ipRateLimiter.RequestCounts = make(map[string]*RequestCounter)
	ipRateLimiter.Blacklist = make(map[string]time.Time)
}

// RateLimitMiddleware returns an Echo middleware for rate limiting
func RateLimitMiddleware(cfg RequestLimiterConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()

			ipRateLimiter.Lock()
			defer ipRateLimiter.Unlock()

			// Clean up the blacklist if the block time has expired
			if blockTime, exists := ipRateLimiter.Blacklist[ip]; exists {
				if time.Now().After(blockTime) {
					delete(ipRateLimiter.Blacklist, ip)
				} else {
					return c.String(http.StatusForbidden, "Access blocked due to unusual activity")
				}
			}

			// Process request count for IP
			counter, found := ipRateLimiter.RequestCounts[ip]
			if !found {
				counter = &RequestCounter{Count: 1, Since: time.Now()}
				ipRateLimiter.RequestCounts[ip] = counter
			} else {
				if time.Since(counter.Since) <= cfg.Interval {
					ipRateLimiter.RequestCounts[ip].Count++
					// counter.Count++
				} else {
					// Reset counter after interval has passed
					ipRateLimiter.RequestCounts[ip].Count = 1
					// counter.Count = 1

					ipRateLimiter.RequestCounts[ip].Since = time.Now()
					// counter.Since = time.Now()
				}
			}

			// Check request rate
			if found {
				if ipRateLimiter.RequestCounts[ip].Count > cfg.RequestLimit {
					ipRateLimiter.Blacklist[ip] = time.Now().Add(cfg.BlockTime)
					return c.String(http.StatusTooManyRequests, "Rate limit exceeded")
				}
			}

			return next(c)
		}
	}
}
