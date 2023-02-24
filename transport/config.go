package transport

import (
	"context"
	"net"
	"net/netip"

	"github.com/darvaza-proxy/slog"
	"github.com/darvaza-proxy/slog/handlers/discard"
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

	// Context
	Context context.Context
	// Logger is the optional logger to record events
	Logger slog.Logger
}

// revive:disable:cyclomatic
// revive:disable:cognitive-complexity

// SetDefaults attempts to fill any configuration gap, specially
// the IP Addresses when interfaces are provided instead
func (cfg *Config) SetDefaults() error {
	// BindAddress, maybe via BindInterface
	if len(cfg.BindAddress) == 0 {
		var addrs []string
		var err error

		if len(cfg.BindInterface) > 0 {
			// All addresses of given interfaces
			addrs, err = GetStringIPAddresses(cfg.BindInterface...)
			if len(addrs) == 0 && err != nil {
				// Error and no address, no luck
				return err
			}
		}

		if len(addrs) == 0 {
			cfg.BindAddress = []string{"0"}
		}
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

// Addresses returns the BindAddress list parsed into net.IP
func (cfg *Config) Addresses() ([]net.IP, error) {
	n := len(cfg.BindAddress)
	out := make([]net.IP, n)

	for i, s := range cfg.BindAddress {
		var ip net.IP

		addr, err := netip.ParseAddr(s)
		if err != nil {
			return out, err
		}

		addr = addr.Unmap()
		if addr.Is4() {
			a4 := addr.As4()
			ip = a4[:]
		} else {
			a16 := addr.As16()
			ip = a16[:]
		}

		out[i] = ip
	}

	return out, nil
}
