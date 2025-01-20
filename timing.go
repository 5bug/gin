// Copyright 2024 Gin Core Team. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin

import (
	"fmt"
	"io"
	"time"
)

// TimingConfig defines the config for Timing middleware.
type TimingConfig struct {
	// Writer is the destination for timing logs.
	// Default: gin.DefaultWriter
	Writer io.Writer

	// Format is a function to format the timing output.
	// Default: DefaultTimingFormatter
	Format TimingFormatter

	// Skip defines a function to skip timing based on the request.
	// Optional.
	Skip func(*Context) bool
}

// TimingFormatter gives the signature of the formatter function passed to TimingWithFormatter.
type TimingFormatter func(params TimingParams) string

// TimingParams defines parameters passed to the formatter.
type TimingParams struct {
	// Request path
	Path string
	// Request method
	Method string
	// Request duration
	Duration time.Duration
	// TimeStamp shows the time after the request.
	TimeStamp time.Time
}

// DefaultTimingFormatter is the default formatter for timing middleware.
var DefaultTimingFormatter = func(param TimingParams) string {
	return fmt.Sprintf("[Timing] %v | %s %s | Duration: %v\n",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		param.Method,
		param.Path,
		param.Duration,
	)
}

// Timing returns a middleware that records the duration of each request.
func Timing() HandlerFunc {
	return TimingWithConfig(TimingConfig{})
}

// TimingWithFormatter instance a Timing middleware with the specified log format function.
func TimingWithFormatter(f TimingFormatter) HandlerFunc {
	return TimingWithConfig(TimingConfig{
		Format: f,
	})
}

// TimingWithWriter instance a Timing middleware with the specified writer.
func TimingWithWriter(out io.Writer) HandlerFunc {
	return TimingWithConfig(TimingConfig{
		Writer: out,
	})
}

// TimingWithConfig instance a Timing middleware with config.
func TimingWithConfig(conf TimingConfig) HandlerFunc {
	formatter := conf.Format
	if formatter == nil {
		formatter = DefaultTimingFormatter
	}

	out := conf.Writer
	if out == nil {
		out = DefaultWriter
	}

	return func(c *Context) {
		// Skip timing if the request matches the skip function
		if conf.Skip != nil && conf.Skip(c) {
			c.Next()
			return
		}

		// Record start time
		start := time.Now()

		// Process request
		c.Next()

		// Calculate duration after handler
		param := TimingParams{
			Path:      c.Request.URL.Path,
			Method:    c.Request.Method,
			TimeStamp: time.Now(),
			Duration:  time.Since(start),
		}

		// Output timing information
		fmt.Fprint(out, formatter(param))
	}
}
