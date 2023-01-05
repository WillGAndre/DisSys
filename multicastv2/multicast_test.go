package multicastv2

import (
	"fmt"
	"testing"
	"time"
)

func TestPartialMulticast(t *testing.T) {
	p1 := NewPeer(4444, false)
	p2 := NewPeer(4445, false)
	p3 := NewPeer(4446, false)

	p1.PingPeer(int(p2.Port), "hello")
	p1.PingPeer(int(p3.Port), "hello")
	p2.PingPeer(int(p3.Port), "hello")
	for {
		fmt.Printf("\n%+v\n", p1)
		fmt.Printf("\n%+v\n", p2)
		fmt.Printf("\n%+v\n", p3)

		time.Sleep(5 * time.Second)
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
