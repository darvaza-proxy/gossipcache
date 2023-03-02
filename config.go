package gossipcache

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/netip"
	"net/url"

	"darvaza.org/core"
	"darvaza.org/slog"
	"darvaza.org/slog/handlers/discard"
	"github.com/hashicorp/memberlist"

	"darvaza.org/gossipcache/transport"
)

const (
	// DefaultCacheBasePath is the default Path under the CacheBaseURL for the groupcache
	DefaultCacheBasePath = "/_groupcache/"
	// DefaultCacheReplicas indicated the default number of replicas for the groupcache
	DefaultCacheReplicas = 50
)

// Config represents the configuration used to setup the GossipCache node
type Config struct {
	// Context
	Context context.Context

	// Logger is an optional logger to bind to memberlist and groupcache
	Logger slog.Logger

	// Transport is the Cluster's transport configuration
	Transport *transport.Config

	// Memberlist is the Cluster's configuration
	Memberlist *memberlist.Config

	// CacheReplicas is the number of replicas in the groupcache.
	// If zero or negative it will be set to DefaultCacheReplicas
	CacheReplicas int

	// CacheBaseURL is the base of the advertised URL for the groupcache.
	// Path and Query string components will be ignored
	CacheBaseURL string
	// CacheBasePath is the path under the CacheBaseURL where the groupcache
	// handler is mounted
	CacheBasePath string

	// ClientTLSConfig is the tls.Config to be used when connecting to other
	// nodes of the cluster when https scheme is used
	ClientTLSConfig *tls.Config
}

// revive:disable:cyclomatic
// revive:disable:cognitive-complexity

// SetDefaults populates Config with defaults, except for CacheBaseURL
// which we will attempt to populate after memberlist is prepared
func (conf *Config) SetDefaults() error {
	// revive:enable:cyclomatic
	// revive:enable:cognitive-complexity

	// Context
	if conf.Context == nil {
		conf.Context = context.Background()
	}

	// Logger
	if conf.Logger == nil {
		conf.Logger = discard.New()
	}

	// CacheReplicas
	if conf.CacheReplicas <= 0 {
		conf.CacheReplicas = DefaultCacheReplicas
	}

	// CacheBaseURL
	if conf.CacheBaseURL != "" {
		s, err := PrepareCacheBaseURL(conf.CacheBaseURL)
		if err != nil {
			// failed to sanitise
			return core.Wrapf(err, "%s: %s: %s", "Config", "CacheBaseURL", "invalid")
		}

		conf.CacheBaseURL = s
	}

	// CacheBasePath
	if s := conf.CacheBasePath; s == "" {
		conf.CacheBasePath = DefaultCacheBasePath
	} else if s[0] != '/' {
		return fmt.Errorf("%s: %s: %s", "Config", "CacheBasePath", "invalid")
	}

	return nil
}

// revive:disable:cyclomatic
// revive:disable:cognitive-complexity

// PrepareCacheBaseURL sanitises a CacheBaseURL value
func PrepareCacheBaseURL(s string) (string, error) {
	// revive:enable:cyclomatic
	// revive:enable:cognitive-complexity

	var u *url.URL
	var err error

	if s == "" {
		// let's go with defaults
		u = &url.URL{}
	} else if u, err = url.Parse(s); err != nil {
		// parse error
		return "", err
	}

	if u.Scheme == "" {
		u.Scheme = "https"
	} else if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("invalid scheme (%q)", u.Scheme)
	}

	if u.Host == "" {
		// pick the first non-localhost address
		ifaces, err := core.GetInterfacesNames("lo")
		if err != nil {
			return "", err
		}

		addrs, err := core.GetStringIPAddresses(ifaces...)
		if err != nil {
			return "", err
		}

		if len(addrs) > 0 {
			u.Host = addrs[0]
		} else {
			u.Host = "0"
		}
	} else if host, port, err := net.SplitHostPort(u.Host); err != nil {
		return "", core.Wrapf(err, "invalid Host (%q)", u.Host)
	} else if port == "80" && u.Scheme == "http" {
		u.Host = host
	} else if port == "443" && u.Scheme == "https" {
		u.Host = host
	}

	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""

	return u.String(), nil
}

// InferCacheBaseURL produces a CacheBaseURL pointing to https on 443/tcp of
// Transport's AdveriseAddr
func InferCacheBaseURL(cfg *memberlist.Config) (string, error) {
	var s string

	ip, _, err := cfg.Transport.FinalAdvertiseAddr(cfg.AdvertiseAddr, 0)
	if err != nil {
		return "", err
	}
	addr, _ := netip.AddrFromSlice(ip)
	addr.Unmap()
	addr.WithZone("")

	if addr.Is6() {
		s = fmt.Sprintf("[%s]", addr.String())
	} else {
		s = addr.String()
	}

	return "https://" + s, nil
}
