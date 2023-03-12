package transport

import (
	"errors"
	"fmt"
	"net"

	"github.com/darvaza-proxy/core"
)

var (
	errBadSet = errors.New("invalid listener set")
)

// Listeners contains externally prepared listeners for our Transport
type Listeners struct {
	TCP []*net.TCPListener
	UDP []*net.UDPConn
}

// Validate checks if the listeners are suitable, and returns
// the addresses and port used
func (lsn *Listeners) Validate() ([]string, int, error) {
	var port int
	var addrs []string
	var err error

	if lsn == nil || len(lsn.TCP) == 0 || len(lsn.UDP) == 0 ||
		len(lsn.TCP) != len(lsn.UDP) {
		return nil, 0, errBadSet
	}

	for i := range lsn.TCP {
		tcp := lsn.TCP[i].Addr().(*net.TCPAddr)
		udp := lsn.UDP[i].LocalAddr().(*net.UDPAddr)

		if port == 0 {
			port = tcp.Port
		}

		err = validatePair(tcp, udp, port)
		if err != nil {
			break
		}

		addrs = append(addrs, tcp.IP.String())
	}

	return addrs, port, err
}

func validatePair(tcp *net.TCPAddr, udp *net.UDPAddr, port int) error {
	var err error

	if port == 0 {
		err = fmt.Errorf("invalid port: %s", tcp.String())
	} else if !tcp.IP.Equal(udp.IP) || tcp.Port != udp.Port || tcp.Port != port {
		err = core.Wrapf(errBadSet, "tcp:%s ≠ udp:%s (port:%v)",
			tcp.String(), udp.String(), port)
	}

	return err
}

// Close closes all listeners
func (lsn *Listeners) Close() error {
	for _, lsn := range lsn.TCP {
		_ = lsn.Close()
	}
	for _, lsn := range lsn.UDP {
		_ = lsn.Close()
	}
	return nil
}

func newListeners(config *Config) (*Listeners, error) {
	// parse BindAddr
	addrs, err := config.Addresses()
	if err != nil {
		return nil, err
	}

	// listen ports
	lsn := &Listeners{}
	if n, err := lsn.listenConfig(addrs, config); err != nil {
		return nil, err
	} else if n < 1 {
		return nil, errors.New("no listening ports")
	} else {
		return lsn, nil
	}
}

// listenConfig attempts to set up listeners on all addresses based
// on the nuances of a Config
func (lsn *Listeners) listenConfig(addrs []net.IP, config *Config) (int, error) {
	givenPort := config.BindPort

	// when not strict, we try to multiple ports
	if !config.BindPortStrict {
		var err error

		for i := 0; i < config.BindPortRetry; i++ {
			var count int

			count, err = lsn.tryListen(i, addrs, givenPort, config)
			if err == nil {
				// success
				return count, nil
			}
		}

		// no luck
		return 0, err
	}

	// but when strict, just once
	return lsn.tryListen(0, addrs, givenPort, config)
}

// tryListen will try to setup listeners considering the attempt count.
// when the port is 0, the first time it will try the default port, and
// go random after that
// when the port is non-zero, it will increment its value on each pass
func (lsn *Listeners) tryListen(pass int, addrs []net.IP, port int, config *Config) (int, error) {
	if port == 0 {
		// the first time we try the default, on the next we
		// go random
		if pass == 0 {
			port = DefaultPort
		}
	} else {
		// if port is fixed, we increment on each pass
		port = port + pass
	}

	return lsn.listen(addrs, port, config)
}

// revive:disable:cognitive-complexity

// listen attempts to listen all addresses on a given port,
// and on success the listeners are stored on the Transport
func (lsn *Listeners) listen(addrs []net.IP, port int, config *Config) (int, error) {
	// revive:enable:cognitive-complexity
	var ok bool

	n := len(addrs)
	tcpListeners := make([]*net.TCPListener, 0, n)
	udpListeners := make([]*net.UDPConn, 0, n)

	// close any success when failing once
	defer func() {
		if !ok {
			for _, tcpLn := range tcpListeners {
				_ = tcpLn.Close()
			}
			for _, udpLn := range udpListeners {
				_ = udpLn.Close()
			}
		}
	}()

	for _, ip := range addrs {
		// TCP
		tcpAddr := &net.TCPAddr{IP: ip, Port: port}
		tcpLn, err := config.ListenTCP("tcp", tcpAddr)
		if err != nil {
			return -1, err
		}

		// appended early so they get closed on error
		tcpListeners = append(tcpListeners, tcpLn)

		if port == 0 {
			// port was random, now we stick to it
			port = tcpLn.Addr().(*net.TCPAddr).Port
		}

		// UDP
		udpAddr := &net.UDPAddr{IP: ip, Port: port}
		udpLn, err := config.ListenUDP("udp", udpAddr)
		if err != nil {
			return -1, err
		}

		// appended early so they get closed on error
		udpListeners = append(udpListeners, udpLn)

		if _, err = setUDPRecvBuffer(udpLn); err != nil {
			return -1, err
		}
	}

	ok = true
	lsn.TCP = tcpListeners
	lsn.UDP = udpListeners
	return n, nil
}
