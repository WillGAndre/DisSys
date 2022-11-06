package peergossip

import (
	"fmt"
	"testing"
	"time"
)

func TestPeerGossip(t *testing.T) {
	p1 := NewPeer(4444)
	p2 := NewPeer(4445)
	p1.Register(int(p2.Port))

	for {
		fmt.Printf("\n%+v\n", p1)
		fmt.Printf("\n%+v\n", p2)

		time.Sleep(5 * time.Second)
	}
}
