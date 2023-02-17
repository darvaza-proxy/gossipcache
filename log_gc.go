package gossipcache

import (
	"github.com/mailgun/groupcache/v2"

	"github.com/darvaza-proxy/slog"
)

var (
	_ groupcache.Logger = (*GroupCacheLogger)(nil)
)

// GroupCacheLogger is a specific log context for groupcache
type GroupCacheLogger struct {
	logger slog.Logger
}

// Printf logs a message under a previously set level and with previously set fields
func (gcl *GroupCacheLogger) Printf(format string, args ...any) {
	gcl.logger.Printf(format, args...)
}

// Error creates a new logger context with level set to Error
func (gcl *GroupCacheLogger) Error() groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger.Error(),
	}
}

// Warn creates a new logger context with level set to Warning
func (gcl *GroupCacheLogger) Warn() groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger.Warn(),
	}
}

// Info creates a new logger context with level set to Info
func (gcl *GroupCacheLogger) Info() groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger.Info(),
	}
}

// Debug creates a new logger context with level set to Debug
func (gcl *GroupCacheLogger) Debug() groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger.Debug(),
	}
}

// ErrorField creates a new logger context with a new field containing an error
func (gcl *GroupCacheLogger) ErrorField(label string, err error) groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger.WithField(label, err),
	}
}

// StringField creates a new logger context with a new field containing a string value
func (gcl *GroupCacheLogger) StringField(label string, val string) groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger.WithField(label, val),
	}
}

// WithFields creates a new logger context with a set of new fields of arbitrary value
func (gcl *GroupCacheLogger) WithFields(fields map[string]any) groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger.WithFields(fields),
	}
}

// NewGroupCacheLogger creates a Logger for groupcache wrapping a given slog.Logger
func NewGroupCacheLogger(l slog.Logger) groupcache.Logger {
	return &GroupCacheLogger{
		logger: l,
	}
}

// SetGroupCacheLogger sets groupcache to use a given slog.Logger
func SetGroupCacheLogger(l slog.Logger) {
	gcl := NewGroupCacheLogger(l)
	groupcache.SetLoggerFromLogger(gcl)
}
