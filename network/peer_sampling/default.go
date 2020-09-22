package peer_sampling

import (
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils"
)

type DefaultPeerSampler struct {
	ValidatorsHandler Handler
}

func (s *DefaultPeerSampler) ShouldConnect(ip utils.IPDesc, id *ids.ShortID) bool {
	return true
}

func (s *DefaultPeerSampler) Connected(p Peer) {
	if s.ValidatorsHandler != nil {
		s.ValidatorsHandler.Connected(p.ID())
	}
}

func (s *DefaultPeerSampler) Disconnected(p Peer) {
	if s.ValidatorsHandler != nil {
		s.ValidatorsHandler.Disconnected(p.ID())
	}
}

func (s *DefaultPeerSampler) Shutdown() {
	// noop
}

func (s *DefaultPeerSampler) PeerListGossipSpacing() time.Duration {
	return time.Minute
}

func (s *DefaultPeerSampler) PeerListGossipSize() int {
	return 100
}
