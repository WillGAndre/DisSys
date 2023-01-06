package multicast

import (
	"fmt"
	"testing"
)

func TestV2MULTI(t *testing.T) {
	p1 := NewPeer(4441, false)
	p2 := NewPeer(4442, false)
	p3 := NewPeer(4443, false)
	p4 := NewPeer(4444, false)
	p5 := NewPeer(4445, true)
	p6 := NewPeer(4446, true)

	p1.PingPeer(int(p1.Port), "hello")
	p1.PingPeer(int(p2.Port), "hello")
	p1.PingPeer(int(p3.Port), "hello")
	p1.PingPeer(int(p4.Port), "hello")
	p1.PingPeer(int(p5.Port), "hello")
	p1.PingPeer(int(p6.Port), "hello")

	p2.PingPeer(int(p2.Port), "hello")
	p2.PingPeer(int(p3.Port), "hello")
	p2.PingPeer(int(p4.Port), "hello")
	p2.PingPeer(int(p5.Port), "hello")
	p2.PingPeer(int(p6.Port), "hello")

	p3.PingPeer(int(p3.Port), "hello")
	p3.PingPeer(int(p4.Port), "hello")
	p3.PingPeer(int(p5.Port), "hello")
	p3.PingPeer(int(p6.Port), "hello")

	p4.PingPeer(int(p4.Port), "hello")
	p4.PingPeer(int(p5.Port), "hello")
	p4.PingPeer(int(p6.Port), "hello")

	p5.PingPeer(int(p5.Port), "hello")
	p5.PingPeer(int(p6.Port), "hello")

	p6.PingPeer(int(p6.Port), "hello")
	for {
	}
}

func TestPartialMulticastWithGold(t *testing.T) {
	p1 := NewPeer(4444, false)
	p2 := NewPeer(4445, false)

	p1.PingPeer(int(p1.Port), "hello")
	p1.PingPeer(int(p2.Port), "hello")
	p2.PingPeer(int(p2.Port), "hello")
	// fmt.Printf("\n%+v\n", p1)
	// fmt.Printf("\n%+v\n", p2)

	p1.PingAll("ping")
	// fmt.Println("____________________________________________")
	fmt.Printf("\nP1 QUEUE: %+v\n", p1.Queue)
	p2.PingAll("ping")
	for {
		// time.Sleep(2 * time.Second)
		// fmt.Printf("\n%+v\n", p1)
		// fmt.Printf("\n%+v\n", p2)
	}
}
