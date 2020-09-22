package peer_sampling

import (
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils"
)

type Sampler interface {
	// Should connect ?
	// Upon hearing of this new peer, should we try connecting to them ?
	// At this point, we cannot assume that we know the ID for sure
	ShouldConnect(ip utils.IPDesc, id *ids.ShortID) bool

	// Called when a connection was successfully achieved
	// to a peer
	Connected(p Peer)

	// Called when a connection to a peer was closed
	Disconnected(p Peer)

	// Shutdown
	Shutdown()

	// Parameters for the gossip exchanges of peer lists
	PeerListGossipSpacing() time.Duration
	PeerListGossipSize() int
}

type Peer interface {
	// Get peer's IP
	IP() utils.IPDesc

	// Get peer's ID
	ID() ids.ShortID

	// Is it an incoming connection
	IsIncoming() bool

	// Close connection
	Close()
}

// Copy this defnition here (from network/handler.go)
// This is so that we can inform the sampler of the Avalance algorithm
// of the content of our view
type Handler interface {
	Connected(id ids.ShortID)
	Disconnected(id ids.ShortID)
}
