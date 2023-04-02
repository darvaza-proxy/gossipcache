package gossipcache

import (
	"log"

	"github.com/hashicorp/memberlist"

	"darvaza.org/slog"
)

// NewMemberlistLogger creates a standard logger to consume
// memberlist logs
func NewMemberlistLogger(l slog.Logger) *log.Logger {
	out := slog.NewLogWriter(l, memberlistLogHandler)
	return log.New(out, "", 0)
}

// SetMemberlistLogger sets a memberlist.Config to use a given slog.Logger
func SetMemberlistLogger(cfg *memberlist.Config, l slog.Logger) error {
	cfg.LogOutput = nil
	cfg.Logger = NewMemberlistLogger(l)
	return nil
}

func memberlistLogHandler(l slog.Logger, s string) error {
	// TODO: parse `s`
	l.Printf("%s", s)
	return nil
}
