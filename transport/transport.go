// Package transport provides a memberlist.Transport implementation
package transport

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"sync/atomic"

	"darvaza.org/core"
	"darvaza.org/slog"
	"github.com/hashicorp/memberlist"
)

var (
	_ memberlist.Transport          = (*Transport)(nil)
	_ memberlist.NodeAwareTransport = (*Transport)(nil)
)

// Transport implements a memberlist.Transport that uses
// an slog.Logger, Cancellable Context, ListenTCP/ListenUDP
// callbacks
type Transport struct {
	wg        core.WaitGroup
	cancel    context.CancelFunc
	cancelled atomic.Bool
	onError   func(err error)
	log       slog.Logger

	tcpListeners []*net.TCPListener
	udpListeners []*net.UDPConn
	streamCh     chan net.Conn
	packetCh     chan *memberlist.Packet
}

// NewWithListeners creates a new transport using preallocated listeners.
// If it fails, it's your responsibility to close them.
// If succeeds, the created transport needs to be explicitly Close()ed
// once it's no longer used
func NewWithListeners(config *Config, lsn *Listeners) (*Transport, error) {
	if config == nil {
		config = &Config{}
	}

	addrs, port, err := lsn.Validate()
	if err != nil {
		return nil, err
	}

	// update config
	config.BindInterface = nil
	config.BindAddress = addrs
	config.BindPort = port

	if err := config.SetDefaults(); err != nil {
		return nil, err
	}

	return newTransport(config, lsn)
}

// New creates a new Transport based on the given configuration
// or defaults.
// If succeeds, the created transport needs to be explicitly Close()ed
// once it's no longer used
func New(config *Config) (*Transport, error) {
	if config == nil {
		config = &Config{}
	}

	if err := config.SetDefaults(); err != nil {
		// bad config
		return nil, err
	}

	return newTransport(config, nil)
}

func newTransport(config *Config, lsn *Listeners) (*Transport, error) {
	ctx, cancel := context.WithCancel(config.Context)

	t := &Transport{
		cancel:  cancel,
		log:     config.Logger,
		onError: config.OnError,

		streamCh: make(chan net.Conn),
		packetCh: make(chan *memberlist.Packet),
	}

	if lsn == nil {
		var err error

		lsn, err = newListeners(config)
		if err != nil {
			return nil, err
		}
	}

	t.tcpListeners = lsn.TCP
	t.udpListeners = lsn.UDP

	t.wg.OnError(func(err error) error {
		var c core.Catcher

		defer t.initiateShutdown()

		c.Try(func() error {
			t.onError(err)
			return nil
		})

		return err
	})

	// and start
	for i := range t.tcpListeners {
		tcpLn := t.tcpListeners[i]
		udpLn := t.udpListeners[i]

		t.wg.Go(func() error {
			return t.tcpLoop(ctx, tcpLn)
		})

		t.wg.Go(func() error {
			return t.udpLoop(ctx, udpLn)
		})
	}

	return t, nil
}

// Shutdown closes the listening ports and
// cancels the workers, and then waits until
// all workers have exited
func (t *Transport) Shutdown() error {
	t.initiateShutdown()

	t.wg.Wait()
	return nil
}

func (t *Transport) initiateShutdown() {
	if t.cancelled.CompareAndSwap(false, true) {
		// stop workers
		t.cancel()

		// close ports
		for i := range t.tcpListeners {
			_ = t.tcpListeners[i].Close()
		}

		for i := range t.udpListeners {
			_ = t.udpListeners[i].Close()
		}
	}
}

// FinalAdvertiseAddr is used by memberlist to find what address and port to
// advertise to other nodes
func (t *Transport) FinalAdvertiseAddr(ip string, port int) (net.IP, int, error) {
	var err error
	var addr netip.Addr

	if ip != "" {
		// use given
		addr, err = core.ParseAddr(ip)
	} else {
		// listener address
		lsnAddr := t.tcpListeners[0]
		if addrPort, ok := core.AddrPort(lsnAddr); !ok {
			err = core.NewUnreachableErrorf(0, nil, "impossible listener address: %v", lsnAddr)
		} else {
			addr = addrPort.Addr()
			port = int(addrPort.Port())
		}
	}

	if err == nil {
		addr, err = t.doFinalAdvertiseAddr(addr)
		if err == nil && addr.IsValid() {
			return addr.AsSlice(), port, nil
		}
	}

	return t.failFinalAdvertiseAddr(err)
}

func (t *Transport) failFinalAdvertiseAddr(err error) (net.IP, int, error) {
	s := "Failed to get IP Address to advertise"
	t.error(err).Print(s)
	if err == nil {
		err = errors.New(s)
	}
	return nil, 0, err
}

func (*Transport) doFinalAdvertiseAddr(addr netip.Addr) (netip.Addr, error) {
	var addrs []netip.Addr
	var err error

	if addr.IsValid() && !addr.IsUnspecified() {
		return addr, nil
	}

	// listening all addresses, pick one
	ifaces, _ := core.GetInterfacesNames("lo")
	if len(ifaces) > 0 {
		addrs, _ = core.GetIPAddresses(ifaces...)
	}
	if len(addrs) == 0 {
		addrs, err = core.GetIPAddresses()
	}

	if len(addrs) > 0 {
		// pick the first
		return addrs[0], nil
	}

	return netip.Addr{}, err
}
