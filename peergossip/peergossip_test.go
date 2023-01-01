package peergossip

import (
	"fmt"
	"testing"
	"time"
)

func TestPartialPeerGossip(t *testing.T) {
	p1 := NewPeer(4444)
	p2 := NewPeer(4445)
	p1.Register(int(p2.Port))

	for {
		fmt.Printf("\n%+v\n", p1)
		fmt.Printf("\n%+v\n", p2)

		time.Sleep(5 * time.Second)
	}
}

func TestFullPeerGossip(t *testing.T) {
	p1 := NewPeer(4444)
	p2 := NewPeer(4445)
	p3 := NewPeer(4446)
	p4 := NewPeer(4447)
	p5 := NewPeer(4448)
	p6 := NewPeer(4449)
	p1.Register(int(p2.Port))
	p2.Register(int(p3.Port))
	p2.Register(int(p4.Port))
	p4.Register(int(p5.Port))
	p4.Register(int(p6.Port))

	for {
		fmt.Printf("\n%+v\n", p1)
		fmt.Printf("\n%+v\n", p2)
		fmt.Printf("\n%+v\n", p3)
		fmt.Printf("\n%+v\n", p4)
		fmt.Printf("\n%+v\n", p5)
		fmt.Printf("\n%+v\n", p6)

		time.Sleep(5 * time.Second)
	}
}
