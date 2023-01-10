package gossipcache

import (
	"net"
	"net/netip"
)

// GetInterfacesNames returns the list of interfaces,
// considering an optional exclusion list
func GetInterfacesNames(except ...string) ([]string, error) {
	var out []string

	s, err := net.Interfaces()
	if err != nil {
		return out, err
	}

	for _, ifi := range s {
		name := ifi.Name

		for _, nope := range except {
			if name == nope {
				name = ""
				break
			}
		}

		if name != "" {
			out = append(out, name)
		}
	}

	return out, nil
}

// GetIPAddress tries to return the address we are most likely to use
// to communicate with the network
func GetIPAddress(ifaces ...string) (netip.Addr, error) {
	// TODO: consider networks and return "best"
	addrs, err := GetIPAddresses(ifaces...)
	if len(addrs) > 0 {
		return addrs[0], nil
	}
	return netip.Addr{}, err
}

// GetIPAddresses returns the list of IP Addresses,
// optionally considering only the given interfaces
func GetIPAddresses(ifaces ...string) ([]netip.Addr, error) {
	var out []netip.Addr

	if len(ifaces) == 0 {
		var err error

		ifaces, err = GetInterfacesNames()
		if err != nil {
			return out, err
		}
	}

	for _, name := range ifaces {
		ifi, err := net.InterfaceByName(name)
		if err != nil {
			return out, err
		}

		addrs, err := ifi.Addrs()
		if err != nil {
			return out, err
		}

		for _, addr := range addrs {
			var s []byte

			switch v := addr.(type) {
			case *net.IPAddr:
				s = v.IP
			case *net.IPNet:
				s = v.IP
			}

			if ip, ok := netip.AddrFromSlice(s); ok {
				out = append(out, ip)
			}
		}
	}

	return out, nil
}
