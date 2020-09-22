package peer_sampling

import (
	"crypto/sha256"
	"encoding/binary"
	"math"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils"
)

type BasaltCostFunction func(seed ids.ShortID, ip utils.IPDesc) uint64

func BasaltUniformCost(seed ids.ShortID, ip utils.IPDesc) uint64 {
	hash := sha256.Sum256(append(seed.ID[:], []byte(ip.String())...))
	return binary.BigEndian.Uint64(hash[:8])
}

func BasaltHierarchicalCost(seed ids.ShortID, ip utils.IPDesc) uint64 {
	if len(ip.IP) < 4 {
		return math.MaxUint64
	}

	if ipv4 := ip.IP.To4(); ipv4 != nil {
		p1h := sha256.Sum256(append(seed.ID[:], ipv4[:1]...))
		p2h := sha256.Sum256(append(seed.ID[:], ipv4[:2]...))
		p3h := sha256.Sum256(append(seed.ID[:], ipv4[:3]...))
		p4h := sha256.Sum256(append(seed.ID[:], []byte(ip.String())...))

		p1 := uint64(binary.BigEndian.Uint16(p1h[:2]))
		p2 := uint64(binary.BigEndian.Uint16(p2h[:2]))
		p3 := uint64(binary.BigEndian.Uint16(p3h[:2]))
		p4 := uint64(binary.BigEndian.Uint16(p4h[:2]))

		return (p1 << 48) | (p2 << 32) | (p3 << 16) | p4
	} else {
		p1h := sha256.Sum256(append(seed.ID[:], ip.IP[:2]...))
		p2h := sha256.Sum256(append(seed.ID[:], ip.IP[:3]...))
		p3h := sha256.Sum256(append(seed.ID[:], ip.IP[:4]...))
		p4h := sha256.Sum256(append(seed.ID[:], []byte(ip.String())...))

		p1 := uint64(binary.BigEndian.Uint16(p1h[:2]))
		p2 := uint64(binary.BigEndian.Uint16(p2h[:2]))
		p3 := uint64(binary.BigEndian.Uint16(p3h[:2]))
		p4 := uint64(binary.BigEndian.Uint16(p4h[:2]))

		return (p1 << 48) | (p2 << 32) | (p3 << 16) | p4
	}
}
