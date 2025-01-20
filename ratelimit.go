// Copyright 2024 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin

import (
	"sync"
	"time"
)

// RateLimitConfig defines the config for RateLimit middleware.
type RateLimitConfig struct {
	// Rate is the number of requests allowed per window
	Rate int

	// Window is the time duration for the rate limit window
	Window time.Duration

	// Identify is a function to get the rate limit key from the request
	// Default: ClientIP
	Identify func(*Context) string

	// OnLimit defines the response when rate limit is exceeded
	// Default: Status 429 (Too Many Requests)
	OnLimit func(*Context)

	// Store is the storage backend for rate limit data
	// Default: in-memory storage using sync.Map
	Store RateLimitStore
}

// RateLimitStore defines the interface for rate limit data storage
type RateLimitStore interface {
	// Allow checks if a request is allowed and updates the counter
	Allow(key string, rate int, window time.Duration) bool
}

// memoryStore implements RateLimitStore using in-memory storage
type memoryStore struct {
	sync.RWMutex
	tokens map[string]*tokenBucket
}

// tokenBucket represents a token bucket for rate limiting
type tokenBucket struct {
	tokens    int
	lastCheck time.Time
}

// defaultIdentify returns the client IP as the rate limit key
var defaultIdentify = func(c *Context) string {
	return c.ClientIP()
}

// defaultOnLimit returns 429 Too Many Requests
var defaultOnLimit = func(c *Context) {
	c.AbortWithStatus(429)
}

// newMemoryStore creates a new in-memory store for rate limiting
func newMemoryStore() *memoryStore {
	return &memoryStore{
		tokens: make(map[string]*tokenBucket),
	}
}

// Allow implements RateLimitStore interface
func (s *memoryStore) Allow(key string, rate int, window time.Duration) bool {
	s.Lock()
	defer s.Unlock()

	now := time.Now()
	bucket, exists := s.tokens[key]
	if !exists {
		s.tokens[key] = &tokenBucket{
			tokens:    rate - 1, // consume one token
			lastCheck: now,
		}
		return true
	}

	// Calculate tokens to add based on time elapsed
	elapsed := now.Sub(bucket.lastCheck)
	newTokens := int(float64(elapsed) / float64(window) * float64(rate))

	if newTokens > 0 {
		bucket.tokens = min(bucket.tokens+newTokens, rate)
		bucket.lastCheck = now
	}

	if bucket.tokens <= 0 {
		return false
	}

	bucket.tokens--
	return true
}

// RateLimit returns a middleware that performs rate limiting
func RateLimit() HandlerFunc {
	return RateLimitWithConfig(RateLimitConfig{})
}

// RateLimitWithConfig returns a middleware with custom config for rate limiting
func RateLimitWithConfig(conf RateLimitConfig) HandlerFunc {
	// Set default values
	if conf.Rate <= 0 {
		conf.Rate = 100 // default 100 requests
	}
	if conf.Window <= 0 {
		conf.Window = time.Minute // default 1 minute window
	}
	if conf.Identify == nil {
		conf.Identify = defaultIdentify
	}
	if conf.OnLimit == nil {
		conf.OnLimit = defaultOnLimit
	}
	if conf.Store == nil {
		conf.Store = newMemoryStore()
	}

	return func(c *Context) {
		// Get identifier for this request
		key := conf.Identify(c)

		// Check if request is allowed
		if !conf.Store.Allow(key, conf.Rate, conf.Window) {
			conf.OnLimit(c)
			return
		}

		c.Next()
	}
}

// min returns the smaller of x or y
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
