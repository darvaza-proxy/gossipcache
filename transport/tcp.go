package transport

import (
	"context"
	"net"
	"time"

	"github.com/hashicorp/memberlist"
)

// DialAddressTimeout is used by memberlist to establish a TCP connection to a particular node
func (*Transport) DialAddressTimeout(addr memberlist.Address, timeout time.Duration) (
	net.Conn, error) {
	dialer := net.Dialer{
		Timeout: timeout,
	}
	return dialer.Dial("tcp", addr.Addr)
}

// DialTimeout is used by memberlist to connect to a particular TCP Address
func (t *Transport) DialTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	peer := memberlist.Address{
		Addr: addr,
	}
	return t.DialAddressTimeout(peer, timeout)
}

// StreamCh is used by memberlist to receive incoming TCP connections
func (t *Transport) StreamCh() <-chan net.Conn {
	return t.streamCh
}

// revive:disable:cognitive-complexity

// tcpLoop is the main routine of the TCP listening workers
func (t *Transport) tcpLoop(ctx context.Context, ln *net.TCPListener) error {
	// revive:enable:cognitive-complexity

	// we explicitly close the listener because we could be interrupted
	// by the cancellation of the parent Context instead of Shutdown()
	defer ln.Close()

	const baseDelay = 5 * time.Millisecond
	const maxDelay = 1 * time.Second

	var errorDelay time.Duration

	for {
		conn, err := ln.Accept()

		if err == nil {
			// no error, incoming connection
			errorDelay = 0

			t.debug().
				WithField(ListenerAddrLabel, ln.Addr()).
				WithField(RemoteAddrLabel, conn.RemoteAddr()).
				Print("Connected")

			select {
			case t.streamCh <- conn:
				// continue
			case <-ctx.Done():
				err := ctx.Err()
				// sorry pal, we are cancelled
				t.error(err).
					WithField(ListenerAddrLabel, ln.Addr()).
					WithField(RemoteAddrLabel, conn.RemoteAddr()).
					Print("Connection Terminated")

				_ = conn.Close()
				return err
			}
		} else if t.cancelled.Load() {
			// shutdown in process, ignore error and exit
			return nil
		} else {
			if errorDelay == 0 {
				// first
				errorDelay = baseDelay
			} else {
				// again? double the delay
				errorDelay *= 2
			}

			if errorDelay > maxDelay {
				// that's too far
				errorDelay = maxDelay
			}

			t.error(err).
				WithField(ListenerAddrLabel, ln.Addr()).
				Print("Error accepting TCP connection")

			select {
			case <-time.After(errorDelay):
				// let's wait a bit
			case <-ctx.Done():
				// cancelled
				return ctx.Err()
			}
		}
	}
}
