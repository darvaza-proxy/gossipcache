package transport

import "darvaza.org/slog"

const (
	// Subsystem indicates our name on the logs
	Subsystem = "gossiptransport"
	// SubsystemLabel is the Field label for the Subsystem
	SubsystemLabel = "subsystem"
	// RemoteAddrLabel is the Field label for the remote party on a connection
	RemoteAddrLabel = "conn"
	// ListenerAddrLabel is the Field label for the local party on a connection
	ListenerAddrLabel = "addr"
	// PacketSizeLabel is the Field label used when logging message size
	PacketSizeLabel = "bytes"
)

func (t *Transport) debug() slog.Logger {
	return t.log.Debug().
		WithField(SubsystemLabel, Subsystem)
}

func (t *Transport) error(err error) slog.Logger {
	l := t.log.Error().
		WithField(SubsystemLabel, Subsystem)

	if err != nil {
		l.WithField(slog.ErrorFieldName, err)
	}
	return l
}
