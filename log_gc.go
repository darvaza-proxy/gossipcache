package gossipcache

import (
	"github.com/mailgun/groupcache/v2"

	"github.com/darvaza-proxy/slog"
)

var (
	_ groupcache.Logger = (*GroupCacheLogger)(nil)
)

type GroupCacheLogLevel int

const (
	GroupCacheErrorLog GroupCacheLogLevel = iota
	GroupCacheWarnLog
	GroupCacheInfoLog
	GroupCacheDebugLog
)

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

func (gcl *GroupCacheLogger) Error() groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger,
		level:  GroupCacheErrorLog,
	}
}

func (gcl *GroupCacheLogger) Warn() groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger,
		level:  GroupCacheWarnLog,
	}
}

func (gcl *GroupCacheLogger) Info() groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger,
		level:  GroupCacheInfoLog,
	}
}

func (gcl *GroupCacheLogger) Debug() groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger,
		level:  GroupCacheDebugLog,
	}
}

func (gcl *GroupCacheLogger) ErrorField(label string, err error) groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger.WithField(label, err),
		level:  gcl.level,
	}
}

func (gcl *GroupCacheLogger) StringField(label string, val string) groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger.WithField(label, val),
		level:  gcl.level,
	}
}

func (gcl *GroupCacheLogger) WithFields(fields map[string]interface{}) groupcache.Logger {
	return &GroupCacheLogger{
		logger: gcl.logger.WithFields(fields),
		level:  gcl.level,
	}
}

func NewGroupCacheLogger(l slog.Logger) groupcache.Logger {
	return &GroupCacheLogger{
		logger: l,
		level:  GroupCacheInfoLog,
	}
}

func SetGroupCacheLogger(l slog.Logger) {
	gcl := NewGroupCacheLogger(l)
	groupcache.SetLoggerFromLogger(gcl)
}
