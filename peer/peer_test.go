package peer

import (
	"testing"
)

func TestSelfPeerPing(t *testing.T) {
	p := NewPeer(4443, 4443, 1)
	PeerShell(p)
}

func TestPeerPing(t *testing.T) {
	p1 := NewPeer(4444, 4445, 0)
	NewPeer(4445, 4446, 0)
	NewPeer(4446, 4447, 0)
	NewPeer(4447, 4444, 0)
	p1.Bind()
}
