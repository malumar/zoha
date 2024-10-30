package main

import (
	"fmt"
	"sync"
	"time"
)

const Minuta = 1000 * 60

// If you want to have second, minute, quarter-hour, hourly and 24-hour counters, you run 5 counters:
// NewLateLimit(1, 1000) // 1 per second
// NewRateLimit(10, Minuta) // 10 per minute
// NewRateLimit(20, Minuta*15) // 20 per quarter of an hour
// NewLateLimit(30, Minuta*60) // per hour
// NewLateLimit(100, Minuta*1440) // per 24h

func main() {
	// how many events (limit) per interval (in ms, 1000 = second)
	rl := NewRateLimiter(4, 1000)
	last := time.Now().UnixNano() / int64(time.Millisecond)
	for i := 0; i < 20; i++ {
		// you need to slow down the program because it runs too fast
		time.Sleep(1 * time.Millisecond)
		now := time.Now().UnixNano() / int64(time.Millisecond)
		lastTime, allow := rl.Add()
		fmt.Println(i, "last time:", lastTime, "allow", allow, now-last)
		last = now
	}
}

func NewRateLimiter(limit int, interval int64) RateLimiter {
	return RateLimiter{
		limit:         float64(limit),
		interval:      float64(interval),
		available:     0,
		lastTimeStamp: 0, // time.Now().UnixNano() / int64(time.Millisecond),
	}
}

// Add /*
func (r *RateLimiter) Add() (int64, bool) {
	// the question is whether it needs to be divided by milliseconds (because we want the result in milliseconds)
	now := time.Now().UnixNano() / int64(time.Millisecond)
	r.lock.RLock()
	defer r.lock.RUnlock()
	// how many ms have passed since the last time
	lastTime := now - r.lastTimeStamp
	r.available += float64(lastTime) * 1.0 / r.interval * r.limit
	if r.available > r.limit {
		r.available = r.limit
	}

	if r.available < 1 {
		return lastTime, false
	}
	r.available -= 1
	r.lastTimeStamp = now
	return lastTime, true
}

type RateLimiter struct {
	lock          sync.RWMutex
	limit         float64
	available     float64
	interval      float64
	lastTimeStamp int64
}
