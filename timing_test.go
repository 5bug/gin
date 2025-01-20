// Copyright 2024 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTiming(t *testing.T) {
	buffer := new(strings.Builder)
	router := New()
	router.Use(TimingWithWriter(buffer))

	router.GET("/test", func(c *Context) {
		time.Sleep(10 * time.Millisecond) // Simulate work
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Contains(t, buffer.String(), "[Timing]")
	assert.Contains(t, buffer.String(), "GET /test")
	assert.Contains(t, buffer.String(), "Duration:")
}

func TestTimingWithFormatter(t *testing.T) {
	buffer := new(strings.Builder)
	router := New()

	router.Use(TimingWithConfig(TimingConfig{
		Writer: buffer,
		Format: func(param TimingParams) string {
			return fmt.Sprintf("CUSTOM %s %s %v\n",
				param.Method,
				param.Path,
				param.Duration,
			)
		},
	}))

	router.GET("/test", func(c *Context) {
		c.String(http.StatusOK, "ok")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w, req)

	assert.Contains(t, buffer.String(), "CUSTOM")
	assert.Contains(t, buffer.String(), "GET /test")
}

func TestTimingWithSkipper(t *testing.T) {
	buffer := new(strings.Builder)
	router := New()

	router.Use(TimingWithConfig(TimingConfig{
		Writer: buffer,
		Skip: func(c *Context) bool {
			return c.Request.URL.Path == "/skip"
		},
	}))

	router.GET("/test", func(c *Context) {
		c.String(http.StatusOK, "ok")
	})

	router.GET("/skip", func(c *Context) {
		c.String(http.StatusOK, "ok")
	})

	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	router.ServeHTTP(w1, req1)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/skip", nil)
	router.ServeHTTP(w2, req2)

	assert.Contains(t, buffer.String(), "/test")
	assert.NotContains(t, buffer.String(), "/skip")
}
