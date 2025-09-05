/*
Copyright © 2025 Furkan Pehlivan furkanpehlivan34@gmail.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"k8s.io/klog/v2"
)

type Logger struct {
	*slog.Logger
}

func NewLogger(format, level string, writer io.Writer) *Logger {
	ho := slog.HandlerOptions{Level: leveler(level)}

	var l *slog.Logger

	switch format {
	case "json":
		handler := slog.NewJSONHandler(writer, &ho)
		l = slog.New(handler)
	default:
		handler := slog.NewTextHandler(writer, &ho)
		l = slog.New(handler)
	}

	klog.SetSlogLogger(l)

	return &Logger{
		l,
	}
}

func leveler(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (l *Logger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		logLevel := slog.LevelInfo
		if rw.statusCode >= 500 {
			logLevel = slog.LevelError
		} else if rw.statusCode >= 400 && rw.statusCode < 500 {
			logLevel = slog.LevelWarn
		}

		l.LogAttrs(
			context.Background(),
			logLevel,
			fmt.Sprintf("%d %s %s %.1fms", rw.statusCode, r.Method, r.URL.String(), duration.Seconds()*1e3),
			slog.Int("status", rw.statusCode),
			slog.String("method", r.Method),
			slog.String("uri", r.URL.String()),
			slog.Float64("duration_ms", duration.Seconds()*1e3),
			slog.String("user_agent", r.UserAgent()),
		)
	})
}

// responseWriter is a wrapper around http.ResponseWriter that allows us to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before writing it to the underlying ResponseWriter.
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
