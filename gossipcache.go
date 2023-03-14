// Package gossipcache provides a Gossip powered groupcache
package gossipcache

import (
	"context"
	"crypto/tls"
	"net/http"
	"sync"
	"time"

	"darvaza.org/cache"
	"darvaza.org/cache/x/groupcache"
	"darvaza.org/core"
	"github.com/hashicorp/memberlist"
)

var (
	_ cache.Store = (*GossipCache)(nil)
)

// GossipCache is a groupcache cluster managed using memberlist
type GossipCache struct {
	*groupcache.HTTPPool

	cancelOnce sync.Once
	cancel     func()

	cluster *Cluster
}

// revive:disable:cognitive-complexity

// NewGossipCacheCluster creates a new Cluster to be used for GossipCache.
// User provided ClusterConfigOtions are applied first, so ours take precedence
func NewGossipCacheCluster(conf *Config, options ...ClusterConfigOption) (*GossipCache, error) {
	// revive:enable:cognitive-complexity

	var gc GossipCache

	// Prepare configuration
	if conf == nil {
		conf = &Config{}
	}

	if err := conf.SetDefaults(); err != nil {
		return nil, err
	}

	// Context
	ctx, cancel := context.WithCancel(conf.Context)
	gc.cancel = cancel

	// Logger
	if logger := conf.Logger; logger != nil {
		groupcache.SetLogger(logger) // unfortunatelly it's global
		options = append(options, WithGossipLogger(logger))
	}

	// Memberlist callbacks
	options = append(options, WithEventDelegate(gc.nodeEvent))
	options = append(options, WithNodeMetaDelegate(func(_ *Cluster, limit int) []byte {
		// Use NodeMeta to advertise our CacheBaseURL
		return []byte(conf.CacheBaseURL)
	}))

	// Prepare cluster
	cluster, cfg, err := Prepare(conf.Memberlist, options...)
	if err != nil {
		// failed
		return nil, err
	}
	gc.cluster = cluster
	conf.Memberlist = cfg

	// Memberlist.Transport
	if conf.Memberlist.Transport == nil {
		t, cfg, err := NewGossipTransport(conf)
		if err != nil {
			// failed
			return nil, err
		}
		conf.Transport = cfg
		conf.Memberlist.Transport = t
	}

	// Finish preparing the cache's configuration
	if conf.CacheBaseURL == "" {
		s, err := InferCacheBaseURL(conf.Memberlist)
		if err != nil {
			return nil, err
		}
		conf.CacheBaseURL = s
	}

	// Create Memberlist cluster
	if err := gc.cluster.Create(); err != nil {
		// failed
		return nil, err
	}

	// Create Groupcache HTTPPool
	gc.HTTPPool = groupcache.NewHTTPPoolOpts(conf.CacheBaseURL, &groupcache.HTTPPoolOptions{
		BasePath: conf.CacheBasePath,
		Replicas: conf.CacheReplicas,
		Transport: func(ctx context.Context) http.RoundTripper {
			// TODO: handle cancel()
			return gc.getClient(ctx, conf.ClientTLSConfig)
		},
		Context: func(req *http.Request) context.Context {
			// TODO: handle cancel()
			return gc.getRequestContext(ctx, req)
		},
	})
	return &gc, nil
}

// Memberlist.EventsDelegate
func (*GossipCache) nodeEvent(_ *Cluster, _ *memberlist.Node, ev memberlist.NodeEventType) {
	switch ev {
	case memberlist.NodeJoin:
	case memberlist.NodeLeave:
	case memberlist.NodeUpdate:
	default:
		// eh?
	}

	panic(core.NewPanicErrorf(2, "NodeEvent:%v: NotImplemented", ev))
}

// http.Handler
func (gc *GossipCache) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// TODO: mTLS
	// TODO: Authorization Token
	gc.HTTPPool.ServeHTTP(rw, req)
}

// http.Client
func (*GossipCache) getClient(_ context.Context, _ *tls.Config) http.RoundTripper {
	// ?
	panic(core.NewPanicErrorf(2, "NotImplemented"))
}

// http.Request.Context
func (*GossipCache) getRequestContext(parent context.Context, _ *http.Request) context.Context {
	// ?
	return parent
}

// Leave cancels groupcache workers and leaves the cluster
func (gc *GossipCache) Leave(timeout time.Duration) error {
	gc.cancelOnce.Do(gc.cancel)

	return gc.cluster.members.Leave(timeout)
}
