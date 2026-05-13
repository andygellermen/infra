package auth

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	entries map[string][]time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit <= 0 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimiter{
		limit:   limit,
		window:  window,
		entries: make(map[string][]time.Time),
	}
}

func (l *RateLimiter) Allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := now.Add(-l.window)
	records := l.entries[key]
	filtered := records[:0]
	for _, record := range records {
		if record.After(cutoff) || record.Equal(cutoff) {
			filtered = append(filtered, record)
		}
	}

	if len(filtered) >= l.limit {
		l.entries[key] = filtered
		return false
	}

	filtered = append(filtered, now)
	l.entries[key] = filtered
	return true
}
