package gossipcache

import (
	"encoding/base64"
	"os"
	"time"

	"github.com/darvaza-proxy/slog"
	"github.com/hashicorp/memberlist"
	"github.com/pkg/errors"
)

var (
	_ ClusterConfigOption = WithAddress("")
	_ ClusterConfigOption = WithGossipPort(0)
	_ ClusterConfigOption = WithGossipAdvertise("", 0)
	_ ClusterConfigOption = WithGossipLogger(nil)
	_ ClusterConfigOption = WithGossipKey("", []byte{})
	_ ClusterConfigOption = WithGossipKeyBase64("", "")

	_ ClusterConfigOption = WithDefaultLANConfig()
	_ ClusterConfigOption = WithDefaultWANConfig()
	_ ClusterConfigOption = WithDefaultLocalConfig()
	_ ClusterConfigOption = WithTransport(nil)

	_ ClusterConfigOption = WithDelegateProtocolVersion(0, 0, 0)
	_ ClusterConfigOption = WithNodeMetaDelegate(nil)
	_ ClusterConfigOption = WithNotifyMsgDelegate(nil)
	_ ClusterConfigOption = WithGetBroadcastDelegate(nil)
	_ ClusterConfigOption = WithLocalStateDelegate(nil)
	_ ClusterConfigOption = WithMergeRemoteStateDelegate(nil)

	_ ClusterConfigOption = WithEventDelegate(nil)
	_ ClusterConfigOption = WithConflictDelegate(nil)
	_ ClusterConfigOption = WithMergeDelegate(nil)
	_ ClusterConfigOption = WithPingDelegate([]byte{}, nil)
	_ ClusterConfigOption = WithAliveDelegate(nil)

	_ memberlist.Delegate         = (*ClusterDelegate)(nil)
	_ memberlist.EventDelegate    = (*ClusterDelegate)(nil)
	_ memberlist.ConflictDelegate = (*ClusterDelegate)(nil)
	_ memberlist.MergeDelegate    = (*ClusterDelegate)(nil)
	_ memberlist.PingDelegate     = (*ClusterDelegate)(nil)
	_ memberlist.AliveDelegate    = (*ClusterDelegate)(nil)
)

// Cluster implements a memberlist for GossipCache
type Cluster struct {
	config  memberlist.Config
	members *memberlist.Memberlist

	delegate ClusterDelegate
}

// ClusterConfigOption is used to adjust the memberlist.Config before creating
// the cluster
type ClusterConfigOption func(*Cluster, *memberlist.Config) error

// WithAddress creates a configuration option to set a specific binding address
func WithAddress(host string) ClusterConfigOption {
	opt := func(_ *Cluster, conf *memberlist.Config) error {
		conf.BindAddr = host
		return nil
	}
	return opt
}

// WithGossipPort creates a configuration option to set a specific TCP/UDP port for
// the Gossip Node
func WithGossipPort(port int16) ClusterConfigOption {
	opt := func(_ *Cluster, conf *memberlist.Config) error {
		conf.BindPort = int(port)
		return nil
	}
	return opt
}

// WithGossipAdvertise creates a configuration option for the Node to advertise different
// address/port than the one it's locally bound to
func WithGossipAdvertise(host string, port int16) ClusterConfigOption {
	opt := func(_ *Cluster, conf *memberlist.Config) error {
		conf.AdvertiseAddr = host
		conf.AdvertisePort = int(port)
		return nil
	}
	return opt
}

// WithGossipLogger creates a configuration option for the Node to use a given slog.Logger
func WithGossipLogger(logger slog.Logger) ClusterConfigOption {
	opt := func(_ *Cluster, conf *memberlist.Config) error {
		return SetMemberlistLogger(conf, logger)
	}
	return opt
}

// WithTransport creates a configuration option to set the cluster's transport
func WithTransport(transport memberlist.Transport) ClusterConfigOption {
	opt := func(_ *Cluster, conf *memberlist.Config) error {
		if transport == nil {
			return errors.New("invalid transport")
		}
		conf.Transport = transport
		return nil
	}
	return opt
}

// WithGossipKeyBase64 creates a configuration option for the cluster's encryption
func WithGossipKeyBase64(salt, key string) ClusterConfigOption {
	bkey, err := base64.RawStdEncoding.DecodeString(key)
	if err != nil {
		// failed to decode key
		err = errors.Wrap(err, "WithGossipKeyBase64")

		return func(_ *Cluster, _ *memberlist.Config) error {
			return err
		}
	}
	return WithGossipKey(salt, bkey)
}

// WithGossipKey creates a configuration option for the cluster's encryption
func WithGossipKey(salt string, key []byte) ClusterConfigOption {
	opt := func(_ *Cluster, conf *memberlist.Config) error {
		if l := len(key); l > 0 {
			// enable encryption
			b := make([]byte, l)
			copy(b, key)
			conf.SecretKey = b

			if len(salt) > 0 {
				conf.Label = salt
			} else {
				conf.Label = ""
			}
			conf.SkipInboundLabelCheck = false
		} else {
			// disable encryption
			conf.SecretKey = []byte{}
		}

		return nil
	}
	return opt
}

// WithDefaultLANConfig sets the cluster configuration to the default
// recommended by memberlist for local network environments
func WithDefaultLANConfig() ClusterConfigOption {
	opt := func(cluster *Cluster, _ *memberlist.Config) error {
		cluster.config = *memberlist.DefaultLANConfig()
		return nil
	}
	return opt
}

// WithDefaultWANConfig sets the cluster configuration to the default
// recommended by memberlist for WAN environments
func WithDefaultWANConfig() ClusterConfigOption {
	opt := func(cluster *Cluster, _ *memberlist.Config) error {
		cluster.config = *memberlist.DefaultWANConfig()
		return nil
	}
	return opt
}

// WithDefaultLocalConfig sets the cluster configuration to the default
// recommended by memberlist for local loopback environments
func WithDefaultLocalConfig() ClusterConfigOption {
	opt := func(cluster *Cluster, _ *memberlist.Config) error {
		cluster.config = *memberlist.DefaultLocalConfig()
		return nil
	}
	return opt
}

// WithDelegateProtocolVersion sets the Protocol versions we handle
func WithDelegateProtocolVersion(version, min, max uint8) ClusterConfigOption {
	opt := func(_ *Cluster, conf *memberlist.Config) error {
		conf.DelegateProtocolVersion = version
		conf.DelegateProtocolMin = min
		conf.DelegateProtocolMax = max
		return nil
	}
	return opt
}

// WithNodeMetaDelegate sets a callback to provide information about the current node
func WithNodeMetaDelegate(handler func(cluster *Cluster, limit int) []byte) ClusterConfigOption {
	opt := func(cluster *Cluster, conf *memberlist.Config) error {
		delegate := &cluster.delegate
		delegate.nodeMeta = handler

		conf.Delegate = delegate
		return nil
	}
	return opt
}

// WithNotifyMsgDelegate sets a callback for received user-data
func WithNotifyMsgDelegate(handler func(cluster *Cluster, userData []byte)) ClusterConfigOption {
	opt := func(cluster *Cluster, conf *memberlist.Config) error {
		delegate := &cluster.delegate
		delegate.notifyMsg = handler

		conf.Delegate = delegate
		return nil
	}
	return opt
}

// WithGetBroadcastDelegate sets a callback to retrieve a list of buffers to broadcast
func WithGetBroadcastDelegate(handler func(*Cluster, int, int) [][]byte) ClusterConfigOption {
	opt := func(cluster *Cluster, conf *memberlist.Config) error {
		delegate := &cluster.delegate
		delegate.getBroadcasts = handler

		conf.Delegate = delegate
		return nil
	}
	return opt
}

// WithLocalStateDelegate sets a callback to send the LocalState to a remote.
// The 'join' boolean indicates this is for a join instead of a push/pull
func WithLocalStateDelegate(handler func(*Cluster, bool) []byte) ClusterConfigOption {
	opt := func(cluster *Cluster, conf *memberlist.Config) error {
		delegate := &cluster.delegate
		delegate.localState = handler

		conf.Delegate = delegate
		return nil
	}
	return opt
}

// WithMergeRemoteStateDelegate sets a callback to receive the LocalState of a remote.
// The 'join' boolean indicates this is for a join instead of a push/pull
func WithMergeRemoteStateDelegate(handler func(*Cluster, []byte, bool)) ClusterConfigOption {
	opt := func(cluster *Cluster, conf *memberlist.Config) error {
		delegate := &cluster.delegate
		delegate.mergeRemoteState = handler

		conf.Delegate = delegate
		return nil
	}
	return opt
}

// WithEventDelegate sets a callback to receive notifications about members joining and leaving
func WithEventDelegate(handler func(*Cluster, *memberlist.Node, memberlist.NodeEventType)) ClusterConfigOption {
	opt := func(cluster *Cluster, conf *memberlist.Config) error {
		delegate := &cluster.delegate
		delegate.event = handler

		conf.Events = delegate
		return nil
	}
	return opt
}

// WithConflictDelegate sets a callback to receive notifications about an attempt to join with a name
// already in use
func WithConflictDelegate(handler func(*Cluster, *memberlist.Node, *memberlist.Node)) ClusterConfigOption {
	opt := func(cluster *Cluster, conf *memberlist.Config) error {
		delegate := &cluster.delegate
		delegate.conflict = handler

		conf.Conflict = delegate
		return nil

	}
	return opt
}

// WithMergeDelegate sets a callback when a merge operation could take place. If
// the return value is non-nil, the merge is canceled.
func WithMergeDelegate(handler func(*Cluster, []*memberlist.Node) error) ClusterConfigOption {
	opt := func(cluster *Cluster, conf *memberlist.Config) error {
		delegate := &cluster.delegate
		delegate.merge = handler

		conf.Merge = delegate
		return nil
	}
	return opt
}

// WithPingDelegate sets a callback to notify an observer how long it took for a ping
// message to complete a round trip.
//
// It can also be used for writing arbitrary byte slices into ack messages
func WithPingDelegate(payload []byte, completed func(*Cluster, *memberlist.Node, time.Duration, []byte)) ClusterConfigOption {
	opt := func(cluster *Cluster, conf *memberlist.Config) error {
		delegate := &cluster.delegate
		delegate.ackPayload = payload
		delegate.pingComplete = completed

		conf.Ping = delegate
		return nil
	}
	return opt
}

// WithAliveDelegate sets a callback to be invoked when a message about a live node is received from the network.
// Returning a non-nil error prevents the node from being considered a peer
func WithAliveDelegate(handler func(*Cluster, *memberlist.Node) error) ClusterConfigOption {
	opt := func(cluster *Cluster, conf *memberlist.Config) error {
		delegate := &cluster.delegate
		delegate.alive = handler

		conf.Alive = delegate
		return nil
	}
	return opt
}

// Prepare prepares a Cluster to be created
func Prepare(conf *memberlist.Config, options ...ClusterConfigOption) (*Cluster, *memberlist.Config, error) {
	var cluster Cluster

	if conf == nil {
		conf = memberlist.DefaultLANConfig()
	}

	cluster.config = *conf              // copy memberlist.Config
	cluster.delegate.cluster = &cluster // bind ClusterDelegate

	for _, opt := range options {
		if opt != nil {
			if err := opt(&cluster, &cluster.config); err != nil {
				// failed to apply configuration option
				return nil, nil, err
			}
		}
	}

	return &cluster, &cluster.config, nil
}

// Create finishes the creation the Cluster. Only to be used after calling Prepare()
func (cluster *Cluster) Create() error {
	if cluster.delegate.cluster != cluster {
		return os.ErrInvalid
	} else if cluster.members != nil {
		return os.ErrExist
	}

	ml, err := memberlist.Create(&cluster.config)
	if err != nil {
		// failed to create cluster
		return err
	}

	// ready
	cluster.members = ml
	return nil
}

// NewCluster creates a new memberlist cluster for GossipCache
func NewCluster(conf *memberlist.Config, options ...ClusterConfigOption) (*Cluster, error) {

	cluster, _, err := Prepare(conf, options...)
	if err != nil {
		// failed to prepare
		return nil, err
	}

	err = cluster.Create()
	if err != nil {
		// failed to create cluster
		return nil, err
	}

	// ready
	return cluster, nil
}

// ClusterDelegate hooks handlers into the Memberlist Cluster life cycle
type ClusterDelegate struct {
	cluster *Cluster

	// Delegate
	nodeMeta         func(*Cluster, int) []byte
	notifyMsg        func(*Cluster, []byte)
	getBroadcasts    func(c *Cluster, overhead, limit int) [][]byte
	localState       func(*Cluster, bool) []byte
	mergeRemoteState func(*Cluster, []byte, bool)

	// EventDelegate
	event func(*Cluster, *memberlist.Node, memberlist.NodeEventType)
	// ConflictDelegate
	conflict func(c *Cluster, existing, other *memberlist.Node)
	// MergeDelegate
	merge func(*Cluster, []*memberlist.Node) error
	// PingDeletegate
	ackPayload   []byte
	pingComplete func(*Cluster, *memberlist.Node, time.Duration, []byte)
	// AliveDelegate
	alive func(*Cluster, *memberlist.Node) error
}

// NodeMeta is used to retrieve meta-data about the current node
// when broadcasting an alive message
func (cd *ClusterDelegate) NodeMeta(limit int) []byte {
	if fn := cd.nodeMeta; fn != nil {
		return fn(cd.cluster, limit)
	}
	return nil
}

// NotifyMsg is called when a user-data message is received.
func (cd *ClusterDelegate) NotifyMsg(userData []byte) {
	if fn := cd.notifyMsg; fn != nil {
		fn(cd.cluster, userData)
	}
}

// GetBroadcasts is called when user data messages can be broadcast.
func (cd *ClusterDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	if fn := cd.getBroadcasts; fn != nil {
		return fn(cd.cluster, overhead, limit)
	}
	return nil
}

// LocalState is used for a TCP Push/Pull
func (cd *ClusterDelegate) LocalState(join bool) []byte {
	if fn := cd.localState; fn != nil {
		return fn(cd.cluster, join)
	}
	return nil
}

// MergeRemoteState is invoked after a TCP Push/Pull
func (cd *ClusterDelegate) MergeRemoteState(buf []byte, join bool) {
	if fn := cd.mergeRemoteState; fn != nil {
		fn(cd.cluster, buf, join)
	}
}

// NotifyJoin is invoked when a node is detected to have joined.
func (cd *ClusterDelegate) NotifyJoin(node *memberlist.Node) {
	if fn := cd.event; fn != nil {
		fn(cd.cluster, node, memberlist.NodeJoin)
	}
}

// NotifyLeave is invoked when a node is detected to have left.
func (cd *ClusterDelegate) NotifyLeave(node *memberlist.Node) {
	if fn := cd.event; fn != nil {
		fn(cd.cluster, node, memberlist.NodeLeave)
	}
}

// NotifyUpdate is invoked when a node is detected to have
func (cd *ClusterDelegate) NotifyUpdate(node *memberlist.Node) {
	if fn := cd.event; fn != nil {
		fn(cd.cluster, node, memberlist.NodeUpdate)
	}
}

// NotifyConflict is invoked when a name conflict is detected
func (cd *ClusterDelegate) NotifyConflict(existing, other *memberlist.Node) {
	if fn := cd.conflict; fn != nil {
		fn(cd.cluster, existing, other)
	}
}

// NotifyMerge is invoked when a merge could take place.
func (cd *ClusterDelegate) NotifyMerge(peers []*memberlist.Node) error {
	if fn := cd.merge; fn != nil {
		return fn(cd.cluster, peers)
	}
	return nil
}

// AckPayload is invoked when an ack is being sent; the returned bytes will be appended to the ack
func (cd *ClusterDelegate) AckPayload() []byte {
	return cd.ackPayload
}

// NotifyPingComplete is invoked when an ack for a ping is received
func (cd *ClusterDelegate) NotifyPingComplete(peer *memberlist.Node, rtt time.Duration, payload []byte) {
	if fn := cd.pingComplete; fn != nil {
		fn(cd.cluster, peer, rtt, payload)
	}
}

// NotifyAlive is invoked when a message about a live node is received from the network
func (cd *ClusterDelegate) NotifyAlive(peer *memberlist.Node) error {
	if fn := cd.alive; fn != nil {
		return fn(cd.cluster, peer)
	}
	return nil
}
