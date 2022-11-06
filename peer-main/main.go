package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	peer "token-ring/peer"
)

func main() {
	args := os.Args[1:]
	port, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalln(err)
	}
	next, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatalln(err)
	}
	lock, err := strconv.Atoi(args[2])
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Starting peer with: %s\n", args)
	p := peer.NewPeer(uint16(port), uint16(next), uint8(lock))
	peer.PeerShell(p)
}
