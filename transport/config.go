package transport

import (
	"context"
	"net"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
)

const (
	// DefaultBindRetry indicates how many times we will try binding a port
	DefaultBindRetry = 4
	// DefaultPort indicates the default TCP/UDP Port to use when zero
	DefaultPort = 7946
)

// Config is the configuration data for Transport
type Config struct {
	// BindInterface is the list of interfaces to listen on
	BindInterface []string
	// BindAddress is the list of addresses to listen on
	BindAddress []string
	// BindPort is the port to listen on, for both TCP and UDP
	BindPort int
	// BindPortStrict tells us not to try other ports
	BindPortStrict bool
	// BindPortRetry indicates how many times we will try finding a port
	BindPortRetry int

	// ListenTCP is the helper to use to listen on a TCP port
	ListenTCP func(network string, laddr *net.TCPAddr) (*net.TCPListener, error)
	// ListenUDP is the helper to use to listen on a UDP port
	ListenUDP func(network string, laddr *net.UDPAddr) (*net.UDPConn, error)

	// OnError is called when a worker returns an error, before initiating
	// a shutdown
	OnError func(error)

	// Context
	Context context.Context
	// Logger is the optional logger to record events
	Logger slog.Logger
}

// revive:disable:cognitive-complexity

// SetDefaults attempts to fill any configuration gap, specially
// the IP Addresses when interfaces are provided instead
func (cfg *Config) SetDefaults() error {
	// revive:enable:cognitive-complexity

	// BindAddress, maybe via BindInterface
	if len(cfg.BindAddress) == 0 {
		addrs, err := cfg.getStringIPAddresses()
		if err != nil {
			return err
		}
		cfg.BindAddress = addrs
	}

	// BindPort
	if cfg.BindPortRetry < 1 {
		cfg.BindPortRetry = DefaultBindRetry
	}

	// Callbacks
	if cfg.ListenTCP == nil {
		cfg.ListenTCP = net.ListenTCP
	}

	if cfg.ListenUDP == nil {
		cfg.ListenUDP = net.ListenUDP
	}

	// Context
	if cfg.Context == nil {
		cfg.Context = context.Background()
	}

	// Logger
	if cfg.Logger == nil {
		cfg.Logger = discard.New()
	}

	return nil
}

func (cfg *Config) getStringIPAddresses() ([]string, error) {
	if len(cfg.BindInterface) > 0 {
		// All addresses of given interfaces
		return core.GetStringIPAddresses(cfg.BindInterface...)
	}

	return []string{"0.0.0.0"}, nil
}

// Addresses returns the BindAddress list parsed into net.IP
func (cfg *Config) Addresses() ([]net.IP, error) {
	n := len(cfg.BindAddress)
	out := make([]net.IP, 0, n)

	for _, s := range cfg.BindAddress {
		ip, err := core.ParseNetIP(s)
		if err != nil {
			return out, err
		}

		out = append(out, ip)
	}

	return out, nil
}
