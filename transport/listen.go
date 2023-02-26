package transport

import "net"

// listenConfig attempts to set up listeners on all addresses based
// on the nuances of a Config
func (t *Transport) listenConfig(addrs []net.IP, config *Config) (int, error) {
	givenPort := config.BindPort

	// when not strict, we try to multiple ports
	if !config.BindPortStrict {
		var err error

		for i := 0; i < config.BindPortRetry; i++ {
			var count int

			count, err = t.tryListen(i, addrs, givenPort, config)
			if err == nil {
				// success
				return count, nil
			}
		}

		// no luck
		return 0, err
	}

	// but when strict, just once
	return t.tryListen(0, addrs, givenPort, config)
}

// tryListen will try to setup listeners considering the attempt count.
// when the port is 0, the first time it will try the default port, and
// go random after that
// when the port is non-zero, it will increment its value on each pass
func (t *Transport) tryListen(pass int, addrs []net.IP, port int, config *Config) (int, error) {
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

	return t.listen(addrs, port, config)
}

// revive:disable:cognitive-complexity

// listen attempts to listen all addresses on a given port,
// and on success the listeners are stored on the Transport
func (t *Transport) listen(addrs []net.IP, port int, config *Config) (int, error) {
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
	t.tcpListeners = tcpListeners
	t.udpListeners = udpListeners
	return n, nil
}
