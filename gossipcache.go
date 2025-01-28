// Package gossipcache provides a Gossip powered groupcache
package gossipcache

import (
	"darvaza.org/cache"
	"darvaza.org/cache/x/groupcache"
)

var (
	_ cache.Store[string] = (*GossipCache)(nil)
)

// GossipCache is a groupcache cluster managed using memberlist
type GossipCache struct {
	*groupcache.HTTPPool
}
