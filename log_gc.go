package gossipcache

import (
	"github.com/mailgun/groupcache/v2"

	"github.com/darvaza-proxy/slog"
)

var (
	_ groupcache.Logger = (*GroupCacheLogger)(nil)
)

// GroupCacheLogLevel indicates what log level groupcache wants to use
// on a log entry
type GroupCacheLogLevel int

const (
	GroupCacheErrorLog GroupCacheLogLevel = iota // GroupCacheErrorLog indicates an Error log entry
	GroupCacheWarnLog                            // GroupCacheWarnLog indicates a Warning log entry
	GroupCacheInfoLog                            // GroupCacheInfoLog indicates an Info log entry
	GroupCacheDebugLog                           // GroupCacheDebugLog indicates a Debug log entry
)

// GroupCacheLogger is a specific log context for groupcache
type GroupCacheLogger struct {
	logger slog.Logger
	level  GroupCacheLogLevel
}

var groupCacheLoggers = []func(slog.Logger, string, ...interface{}){
	GroupCacheErrorLog: func(l slog.Logger, format string, args ...interface{}) {
		l.Error(format, args...)
	},
	GroupCacheWarnLog: func(l slog.Logger, format string, args ...interface{}) {
		l.Warn(format, args...)
	},
	GroupCacheInfoLog: func(l slog.Logger, format string, args ...interface{}) {
		l.Info(format, args...)
	},
	GroupCacheDebugLog: func(l slog.Logger, format string, args ...interface{}) {
		l.Debug(format, args...)
	},
}

// Printf logs a message under a previously set level and with previously set fields
func (gcl *GroupCacheLogger) Printf(format string, args ...interface{}) {
	var fn func(slog.Logger, string, ...interface{})
	var level = int(gcl.level)

	if level >= 0 && level < len(groupCacheLoggers) {
		fn = groupCacheLoggers[gcl.level]
	}
	if fn == nil {
		fn = groupCacheLoggers[GroupCacheErrorLog]
	}
	fn(gcl.logger, format, args...)
}

// Error creates a new logger context with level set to Error
func (gcl *GroupCacheLogger) Error() groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger,
		level:  GroupCacheErrorLog,
	}
}

// Warn creates a new logger context with level set to Warning
func (gcl *GroupCacheLogger) Warn() groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger,
		level:  GroupCacheWarnLog,
	}
}

// Info creates a new logger context with level set to Info
func (gcl *GroupCacheLogger) Info() groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger,
		level:  GroupCacheInfoLog,
	}
}

// Debug creates a new logger context with level set to Debug
func (gcl *GroupCacheLogger) Debug() groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger,
		level:  GroupCacheDebugLog,
	}
}

// ErrorField creates a new logger context with a new field containing an error
func (gcl *GroupCacheLogger) ErrorField(label string, err error) groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger.WithField(label, err),
		level:  gcl.level,
	}
}

// StringField creates a new logger context with a new field containing a string value
func (gcl *GroupCacheLogger) StringField(label string, val string) groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger.WithField(label, val),
		level:  gcl.level,
	}
}

// WithFields creates a new logger context with a set of new fields of arbitrary value
func (gcl *GroupCacheLogger) WithFields(fields map[string]interface{}) groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger.WithFields(fields),
		level:  gcl.level,
	}
}

// NewGroupCacheLogger creates a Logger for groupcache wrapping a given slog.Logger
func NewGroupCacheLogger(l slog.Logger) groupcache.Logger {
	return &GroupCacheLogger{
		logger: l,
		level:  GroupCacheInfoLog,
	}
}

// SetGroupCacheLogger sets groupcache to use a given slog.Logger
func SetGroupCacheLogger(l slog.Logger) {
	gcl := NewGroupCacheLogger(l)
	groupcache.SetLoggerFromLogger(gcl)
}
