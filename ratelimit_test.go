// Copyright 2024 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockStore implements RateLimitStore interface for testing
type mockStore struct {
	sync.Mutex
	allowCall func(key string, rate int, window time.Duration) bool
}

func (m *mockStore) Allow(key string, rate int, window time.Duration) bool {
	m.Lock()
	defer m.Unlock()
	return m.allowCall(key, rate, window)
}

func newMockStore(allowCall func(key string, rate int, window time.Duration) bool) *mockStore {
	return &mockStore{allowCall: allowCall}
}

func TestRateLimit_Basic(t *testing.T) {
	r := New()
	r.Use(RateLimitWithConfig(RateLimitConfig{
		Rate:   2,
		Window: time.Second,
	}))
	r.GET("/test", func(c *Context) {
		c.String(http.StatusOK, "ok")
	})

	// First request should succeed
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should succeed
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:1234"
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Third request should fail (rate limit exceeded)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.1:1234"
	r.ServeHTTP(w3, req3)
	assert.Equal(t, 429, w3.Code)
}

func TestRateLimit_DifferentIPs(t *testing.T) {
	r := New()
	r.Use(RateLimitWithConfig(RateLimitConfig{
		Rate:   1,
		Window: time.Second,
	}))
	r.GET("/test", func(c *Context) {
		c.String(http.StatusOK, "ok")
	})

	// Request from first IP
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:1234"
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Request from second IP should succeed
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:1234"
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Second request from first IP should fail
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.1:1234"
	r.ServeHTTP(w3, req3)
	assert.Equal(t, 429, w3.Code)
}

func TestRateLimit_CustomIdentifier(t *testing.T) {
	r := New()
	r.Use(RateLimitWithConfig(RateLimitConfig{
		Rate:   1,
		Window: time.Second,
		Identify: func(c *Context) string {
			return c.GetHeader("X-API-Key")
		},
	}))
	r.GET("/test", func(c *Context) {
		c.String(http.StatusOK, "ok")
	})

	// Request with first API key
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-API-Key", "key1")
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Request with second API key should succeed
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-API-Key", "key2")
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Second request with first API key should fail
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.Header.Set("X-API-Key", "key1")
	r.ServeHTTP(w3, req3)
	assert.Equal(t, 429, w3.Code)
}

func TestRateLimit_CustomResponse(t *testing.T) {
	r := New()
	r.Use(RateLimitWithConfig(RateLimitConfig{
		Rate:   1,
		Window: time.Second,
		OnLimit: func(c *Context) {
			c.JSON(429, gin.H{
				"error": "too many requests",
				"wait":  "1s",
			})
		},
	}))
	r.GET("/test", func(c *Context) {
		c.String(http.StatusOK, "ok")
	})

	// First request should succeed
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request should fail with custom response
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 429, w2.Code)
	assert.Contains(t, w2.Body.String(), "too many requests")
	assert.Contains(t, w2.Body.String(), "wait")
}
