// Package transport provides a memberlist.Transport implementation
package transport

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"sync"
	"sync/atomic"

	"github.com/darvaza-proxy/slog"
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
	wg        sync.WaitGroup
	cancel    context.CancelFunc
	cancelled atomic.Bool
	log       slog.Logger

	tcpListeners []*net.TCPListener
	udpListeners []*net.UDPConn
	streamCh     chan net.Conn
	packetCh     chan *memberlist.Packet
}

// revive:disable:cognitive-complexity

// New creates a new Transport based on the given configuration
// or defaults.
func New(config *Config) (*Transport, error) {
	if config == nil {
		config = &Config{}
	}

	if err := config.SetDefaults(); err != nil {
		// bad config
		return nil, err
	}

	ctx, cancel := context.WithCancel(config.Context)

	t := &Transport{
		cancel: cancel,
		log:    config.Logger,

		streamCh: make(chan net.Conn),
		packetCh: make(chan *memberlist.Packet),
	}

	// parse BindAddr
	addrs, err := config.Addresses()
	if err != nil {
		return nil, err
	}

	// listen ports
	if n, err := t.listenConfig(addrs, config); err != nil {
		return nil, err
	} else if n < 1 {
		return nil, errors.New("no listening ports")
	}

	// and start
	for i := range t.tcpListeners {
		tcpLn := t.tcpListeners[i]
		udpLn := t.udpListeners[i]

		t.wg.Add(2)

		go func() {
			defer t.wg.Done()
			t.tcpLoop(ctx, tcpLn)
		}()
		go func() {
			defer t.wg.Done()
			t.udpLoop(ctx, udpLn)
		}()
	}

	return t, nil
}

// Shutdown closes the listening ports and
// cancels the workers, and then waits until
// all workers have exited
func (t *Transport) Shutdown() error {
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

	t.wg.Wait()
	return nil
}

// FinalAdvertiseAddr is used by memberlist to find what address and port to
// advertise to other nodes
func (t *Transport) FinalAdvertiseAddr(ip string, port int) (net.IP, int, error) {
	var err error

	if ip != "" {
		var givenAddr net.IP

		// use the given address
		addr, err := netip.ParseAddr(ip)
		if err != nil {
			return nil, 0, err
		}

		addr = addr.Unmap()
		if addr.Is4() {
			a4 := addr.As4()
			givenAddr = a4[:]
		} else {
			a16 := addr.As16()
			givenAddr = a16[:]
		}

		return givenAddr, port, nil
	}

	tcpAddr := t.tcpListeners[0].Addr().(*net.TCPAddr)
	if tcpAddr.IP.IsUnspecified() {
		var addrs []netip.Addr

		// listening all addresses, pick one
		ifaces, _ := GetInterfacesNames("lo")
		if len(ifaces) > 0 {
			addrs, _ = GetIPAddresses(ifaces...)
		}
		if len(addrs) == 0 {
			addrs, err = GetIPAddresses()
		}

		if len(addrs) > 0 {
			addr := addrs[0].Unmap()
			if addr.Is4() {
				a4 := addr.As4()
				tcpAddr.IP = a4[:]
			} else {
				a16 := addr.As16()
				tcpAddr.IP = a16[:]
			}

			err = nil
		} else {
			s := "Failed to get IP Address to advertise"
			t.error(err).Print(s)

			if err == nil {
				err = errors.New(s)
			}
		}
	}

	return tcpAddr.IP, tcpAddr.Port, err
}
