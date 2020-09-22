package peer_sampling

import (
	"crypto/rand"
	"math"
	"sync"
	"time"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils"
	"github.com/ava-labs/avalanchego/utils/logging"
)

// Implementation of the BASALT peer sampling algorithm
// See our paper for details

type BasaltPeerSampler struct {
	ViewSize          int
	SeedRenewInterval time.Duration
	SeedRenewCount    int
	CostFunction      BasaltCostFunction

	BootstrapPeers    []ids.ShortID
	ValidatorsHandler Handler

	log logging.Logger

	seeds      []ids.ShortID
	view       []Peer // nil = no peer yet
	renewIndex int

	incomingNotInView map[Peer]struct{}

	stateLock    sync.Mutex
	mustShutdown bool
}

func (s *BasaltPeerSampler) randomSeed() ids.ShortID {
	var bytes [20]byte
	_, err := rand.Read(bytes[:])
	if err != nil {
		s.log.Fatal("Could not generate crypto rand: %s", err)
	}
	return ids.ShortID{ID: &bytes}
}

func (s *BasaltPeerSampler) Initialize(log logging.Logger) {
	s.log = log
	s.log.Info("Initializing the BASALT peer sampling algorithm")

	s.seeds = make([]ids.ShortID, s.ViewSize)
	for i := range s.seeds {
		s.seeds[i] = s.randomSeed()
	}
	s.view = make([]Peer, s.ViewSize)

	s.incomingNotInView = make(map[Peer]struct{})

	go s.renewSeedsRegularly()
}

func (s *BasaltPeerSampler) ShouldConnect(ip utils.IPDesc, id *ids.ShortID) bool {
	//s.log.Debug("BASALT: ShouldConnect %s", ip.String())

	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	for i := range s.seeds {
		if s.view[i] == nil || s.CostFunction(s.seeds[i], ip) < s.CostFunction(s.seeds[i], s.view[i].IP()) {
			return true
		}
	}
	return false
}

func (s *BasaltPeerSampler) Connected(p Peer) {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	// Special treatement to bootstrap peers: just propagate connection event
	// to validators handler, and ignore them in Basalt
	if s.isBootstrapPeer(p) {
		s.ValidatorsHandler.Connected(p.ID())
		return
	}

	s.log.Debug("BASALT: connected to %s (incoming: %t)", p.IP(), p.IsIncoming())

	prevPeers := make([]Peer, s.ViewSize+1)
	copy(prevPeers[:s.ViewSize], s.view[:])
	prevPeers[s.ViewSize] = p

	for i := range s.view {
		s.updateViewSlot(i, []Peer{p})
	}

	s.closeRemovedPeers(prevPeers)

	for i := range s.view {
		if s.view[i] == p {
			Trace("A %s", p.ID())
			s.ValidatorsHandler.Connected(p.ID())
			break
		}
	}
}

func (s *BasaltPeerSampler) Disconnected(p Peer) {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	// Special treatement to bootstrap peers: just propagate disconnection event
	// to validators handler, and ignore them in Basalt
	if s.isBootstrapPeer(p) {
		s.ValidatorsHandler.Disconnected(p.ID())
		return
	}

	s.log.Debug("BASALT: disconnected from %s (incoming: %t)", p.IP(), p.IsIncoming())

	if _, ok := s.incomingNotInView[p]; ok {
		delete(s.incomingNotInView, p)
	}

	// Replace disconnected peer by best matching peer amongst remaining ones

	// First, remove disconnected peer everywhere
	positions := make([]int, 0, 8)
	for i := range s.view {
		if s.view[i] == p {
			s.view[i] = nil
			positions = append(positions, i)
		}
	}

	// Then add back new best matching peers
	for _, i := range positions {
		s.updateViewSlot(i, s.view)
		s.updateViewSlotFromIncomingNotInView(i)
	}
	s.cleanupIncomingNotInView()

	Trace("R %s", p.ID())
	s.ValidatorsHandler.Disconnected(p.ID())
}

func (s *BasaltPeerSampler) Shutdown() {
	s.stateLock.Lock()
	defer s.stateLock.Unlock()

	s.log.Info("BASALT: Shutdown")

	s.mustShutdown = true
}

func (s *BasaltPeerSampler) renewSeedsRegularly() {
	t := time.NewTicker(s.SeedRenewInterval)
	defer t.Stop()

	for range t.C {
		s.stateLock.Lock()

		if s.mustShutdown {
			s.stateLock.Unlock()
			return
		}

		s.log.Debug("BASALT: Renewing some seeds...")

		prevPeers := make([]Peer, s.ViewSize)
		copy(prevPeers[:], s.view[:])

		for c := 0; c < s.SeedRenewCount; c++ {
			i := s.renewIndex
			s.renewIndex = (s.renewIndex + 1) % s.ViewSize

			s.seeds[i] = s.randomSeed()
			s.updateViewSlot(i, prevPeers)
			s.updateViewSlotFromIncomingNotInView(i)
		}

		s.closeRemovedPeers(prevPeers)
		s.cleanupIncomingNotInView()

		s.stateLock.Unlock()
	}
}

func (s *BasaltPeerSampler) updateViewSlot(i int, candidates []Peer) {
	if len(candidates) == 0 {
		return
	}

	var bestCost uint64
	if s.view[i] == nil {
		bestCost = math.MaxUint64
	} else {
		bestCost = s.CostFunction(s.seeds[i], s.view[i].IP())
	}

	for _, candidate := range candidates {
		if candidate == nil {
			continue
		}

		candidateCost := s.CostFunction(s.seeds[i], candidate.IP())
		if candidateCost < bestCost {
			s.log.Debug("BASALT: best match for slot %d: %s (cost %016X)", i, candidate.IP(), candidateCost)
			s.view[i] = candidate
			bestCost = candidateCost
		}
	}
}

func (s *BasaltPeerSampler) updateViewSlotFromIncomingNotInView(i int) {
	var bestCost uint64
	if s.view[i] == nil {
		bestCost = math.MaxUint64
	} else {
		bestCost = s.CostFunction(s.seeds[i], s.view[i].IP())
	}

	for candidate := range s.incomingNotInView {
		candidateCost := s.CostFunction(s.seeds[i], candidate.IP())
		if candidateCost < bestCost {
			s.log.Debug("BASALT: best match for slot %d: %s (cost %016X), from an old incoming connection", i, candidate.IP(), candidateCost)
			s.view[i] = candidate
			bestCost = candidateCost
		}
	}
}

func (s *BasaltPeerSampler) cleanupIncomingNotInView() {
	for _, peer := range s.view {
		if _, ok := s.incomingNotInView[peer]; ok {
			delete(s.incomingNotInView, peer)
			Trace("A %s", peer.ID())
			s.ValidatorsHandler.Connected(peer.ID())
		}
	}
}

func (s *BasaltPeerSampler) closeRemovedPeers(prevPeers []Peer) {
	for _, peer := range prevPeers {
		if peer == nil {
			continue
		}

		stillInView := false
		for _, viewPeer := range s.view {
			if viewPeer == peer {
				stillInView = true
				break
			}
		}
		if !stillInView && !s.isBootstrapPeer(peer) {
			Trace("R %s", peer.ID())
			s.ValidatorsHandler.Disconnected(peer.ID())

			if !peer.IsIncoming() && false {
				s.log.Debug("BASALT: dropping connection to %s", peer.IP())
				go peer.Close()
			} else {
				s.incomingNotInView[peer] = struct{}{}
			}
		}
	}
}

func (s *BasaltPeerSampler) isBootstrapPeer(peer Peer) bool {
	for _, bp := range s.BootstrapPeers {
		if bp.String() == peer.ID().String() {
			return true
		}
	}
	return false
}

func (s *BasaltPeerSampler) PeerListGossipSpacing() time.Duration {
	return time.Second
}

func (s *BasaltPeerSampler) PeerListGossipSize() int {
	return 4
}
