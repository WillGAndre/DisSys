package multicast

import (
	"fmt"
	"testing"
	"time"
)

func TestPartialMulticast(t *testing.T) {
	p1 := NewPeer(4444)
	p2 := NewPeer(4445)
	// p1.PingPeer(int(p2.Port))

	for {
		fmt.Printf("\n%+v\n", p1)
		fmt.Printf("\n%+v\n", p2)

		time.Sleep(5 * time.Second)
		p1.PingPeer(int(p2.Port))
	}
}

func TestPartialGold(t *testing.T) {
	p1 := NewPeer(4444)
	p2 := NewPeer(4445)
	g1 := NewGoldPeer(4446, 4447)
	g2 := NewGoldPeer(4447, 4446)

	p1.PingPeer(int(p2.Port))
	p1.PingGold(int(g1.Port))
	p1.PingGold(int(g2.Port))
	p2.PingGold(int(g1.Port))
	p2.PingGold(int(g2.Port))

	for {
		fmt.Printf("\n%+v\n", p1)
		fmt.Printf("\n%+v\n", p2)

		fmt.Printf("\n%+v\n", g1)
		fmt.Printf("\n%+v\n", g2)

		time.Sleep(5 * time.Second)
		p1.PingGolds()
		p2.PingGolds()
		p1.PingPeer(int(p2.Port))
	}
}

func TestMulticastProc(t *testing.T) {
	p1 := NewPeer(4444)
	p2 := NewPeer(4445)
	p3 := NewPeer(4448)
	p4 := NewPeer(4449)
	p5 := NewPeer(4451)
	g1 := NewGoldPeer(4446, 4447)
	g2 := NewGoldPeer(4447, 4446)

	// peer recon
	p1.PingPeer(int(p2.Port))
	p1.PingPeer(int(p3.Port))
	p1.PingPeer(int(p4.Port))
	p1.PingPeer(int(p5.Port))

	p2.PingPeer(int(p3.Port))
	p2.PingPeer(int(p4.Port))
	p2.PingPeer(int(p5.Port))

	// gold peer recon
	p1.PingGold(int(g1.Port))
	p1.PingGold(int(g2.Port))
	p2.PingGold(int(g1.Port))
	p2.PingGold(int(g2.Port))
	p3.PingGold(int(g1.Port))
	p3.PingGold(int(g2.Port))
	p4.PingGold(int(g1.Port))
	p4.PingGold(int(g2.Port))
	p5.PingGold(int(g1.Port))
	p5.PingGold(int(g2.Port))

	g1.ExchangeQueue(int(g2.Port))

	for {
		// fmt.Printf("\n%+v\n", p1)
		// fmt.Printf("\n%+v\n", p2)
		fmt.Printf("\n%+v\n", g1)
		fmt.Printf("\n%+v\n", g2)

		time.Sleep(5 * time.Second)
	}
}
