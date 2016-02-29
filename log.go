// Package echologrus provides a middleware for echo that logs request details
// via the iDlogger library
// This is a clone of fknsrs.biz/p/echo-logrus
package iDechoLog

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"go.iondynamics.net/iDlogger"
	"go.iondynamics.net/iDlogger/priority"
)

// New returns a new middleware handler with a default name and logger
func New() echo.MiddlewareFunc {
	return NewWithName("web")
}

// NewWithName returns a new middleware handler with the specified name
func NewWithName(name string) echo.MiddlewareFunc {
	return NewWithNameAndLogger(name, iDlogger.StandardLogger())
}

// NewWithNameAndLogger returns a new middleware handler with the specified name and logger
func NewWithNameAndLogger(name string, l *iDlogger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			event := &iDlogger.Event{
				Logger: l,
				Data: map[string]interface{}{
					"request": c.Request().RequestURI,
					"method":  c.Request().Method,
					"remote":  c.Request().RemoteAddr,
				},
				Time: time.Now(),
				Priority: priority.Informational,
			}

			if reqID := c.Request().Header.Get("X-Request-Id"); reqID != "" {
				event.Data["request_id"] = reqID
			}

			event.Message = "started handling request"

			l.Log(event)

			if err := next(c); err != nil {
				c.Error(err)
			}

			latency := time.Since(start)

			event.Data["status"] = c.Response().Status()
			event.Data["text_status"] = http.StatusText(c.Response().Status())
			event.Data["took"] = latency
			event.Data[fmt.Sprintf("measure#%s.latency", name)] = latency.Nanoseconds()


			event.Message = "completed handling request"
			l.Log(event)

			return nil
		}
	}
}