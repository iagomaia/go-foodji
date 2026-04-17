package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		traceID := ""
		if span := trace.SpanFromContext(c.Request.Context()); span.SpanContext().IsValid() {
			traceID = span.SpanContext().TraceID().String()
		}

		log.InfoContext(c.Request.Context(), "request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.FullPath()),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("latency", time.Since(start)),
			slog.String("trace_id", traceID),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}
