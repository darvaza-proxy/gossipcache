package gossipcache

import (
	"log"

	"github.com/hashicorp/memberlist"

	"github.com/darvaza-proxy/slog"
)

func NewMemberlistLogger(l slog.Logger) *log.Logger {
	out := slog.NewLogWriter(l, memberlistLogHandler)
	return log.New(out, "", 0)
}

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
