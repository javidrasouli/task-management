package logutil

import (
	"context"

	"go.uber.org/zap"
)

type contextKey struct{}

// New returns a configured zap logger. Use env "production" for JSON output;
// any other value gives a development logger with colored console output.
func New(env string) (*zap.Logger, error) {
	if env == "production" || env == "prod" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

// WithContext stores the logger in ctx so downstream layers can retrieve it.
func WithContext(ctx context.Context, log *zap.Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, log)
}

// FromContext retrieves the logger stored by WithContext. Falls back to a
// no-op logger so callers never have to nil-check.
func FromContext(ctx context.Context) *zap.Logger {
	if log, ok := ctx.Value(contextKey{}).(*zap.Logger); ok && log != nil {
		return log
	}
	return zap.NewNop()
}
