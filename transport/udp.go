package transport

import (
	"context"
	"net"
	"time"

	"github.com/hashicorp/memberlist"
)

const (
	udpRecvBufMaxSize = 2 * 1024 * 1024
	udpPacketBufSize  = 64 * 1024
)

// setUDPRecvBuffer attempts to set a large receive buffer to a UDP listener
func setUDPRecvBuffer(udpLn *net.UDPConn) (int, error) {
	var err error

	size := udpRecvBufMaxSize
	for size > 0 {
		if err = udpLn.SetReadBuffer(size); err == nil {
			// success
			return size, nil
		}

		// try smaller
		size = size / 2
	}

	// no luck
	return size, err
}

// WriteToAddress is used by memberlist to send a UDP message to a particular Node
func (t *Transport) WriteToAddress(b []byte, addr memberlist.Address) (time.Time, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", addr.Addr)
	if err != nil {
		return time.Time{}, err
	}

	_, err = t.udpListeners[0].WriteTo(b, udpAddr)
	return time.Now(), err
}

// WriteTo is used by memberlist to send a UDP message to a particular address
func (t *Transport) WriteTo(b []byte, addr string) (time.Time, error) {
	peer := memberlist.Address{
		Addr: addr,
	}
	return t.WriteToAddress(b, peer)
}

// PacketCh is used by memberlist to receive UDP messages
func (t *Transport) PacketCh() <-chan *memberlist.Packet {
	return t.packetCh
}

// revive:disable:cognitive-complexity

// udpLoop is the main routine of the UDP listening workers
func (t *Transport) udpLoop(ctx context.Context, ln *net.UDPConn) {
	// revive:enable:cognitive-complexity

	// we explicitly close the listener because we could be interrupted
	// by the cancellation of the parent Context instead of Shutdown()
	defer ln.Close()

	for {
		buf := make([]byte, udpPacketBufSize)
		n, addr, err := ln.ReadFrom(buf)
		ts := time.Now()

		if err != nil {
			// error
			if t.cancelled.Load() {
				// shutdown in process, ignore error and exit
				return
			}

			t.error(err).
				WithField(ListenerAddrLabel, ln.LocalAddr()).
				Print("Error reading UDP packet")
		} else if n < 1 {
			t.error(nil).
				WithField(ListenerAddrLabel, ln.LocalAddr()).
				WithField(RemoteAddrLabel, addr).
				Print("Empty UDP packet received")
		} else {
			t.debug().
				WithField(ListenerAddrLabel, ln.LocalAddr()).
				WithField(RemoteAddrLabel, addr).
				WithField(PacketSizeLabel, n).
				Print("UDP packet received")

			msg := &memberlist.Packet{
				Buf:       buf[:n],
				From:      addr,
				Timestamp: ts,
			}

			select {
			case t.packetCh <- msg:
				// continue
			case <-ctx.Done():
				// cancelled
				t.error(ctx.Err()).
					WithField(ListenerAddrLabel, ln.LocalAddr()).
					WithField(RemoteAddrLabel, addr).
					WithField(PacketSizeLabel, n).
					Print("UDP packet discarded")
				return
			}
		}
	}
}
