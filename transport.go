package gossipcache

import (
	"darvaza.org/gossipcache/transport"
	"github.com/hashicorp/memberlist"
)

// revive:disable:cognitive-complexity

// NewGossipTransportConfig creates a transport.Config from a Config
func NewGossipTransportConfig(conf *Config) (*transport.Config, error) {
	// revive:enable:cognitive-complexity
	tc := conf.Transport
	if tc == nil {
		tc = &transport.Config{}
	}

	// Context
	if tc.Context == nil {
		tc.Context = conf.Context
	}

	// Logger
	if tc.Logger == nil {
		tc.Logger = conf.Logger
	}

	// BindAddress
	if len(tc.BindAddress) == 0 {
		s := conf.Memberlist.BindAddr
		if s == "" {
			s = "0.0.0.0"
		}
		tc.BindAddress = []string{s}
	}
	// BindPort
	if tc.BindPort == 0 {
		tc.BindPort = conf.Memberlist.BindPort
	}

	if err := tc.SetDefaults(); err != nil {
		return nil, err
	}
	return tc, nil
}

// NewGossipTransport creates a transport.Transport from a Config
func NewGossipTransport(conf *Config) (memberlist.Transport, *transport.Config, error) {
	// Config
	tc, err := NewGossipTransportConfig(conf)
	if err != nil {
		return nil, tc, err
	}

	t, err := transport.New(tc)
	if err != nil {
		return nil, tc, err
	}

	return t, tc, nil
}
