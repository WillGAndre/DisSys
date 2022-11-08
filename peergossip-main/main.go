package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	peergossip "token-ring/peergossip"
)

func main() {
	args := os.Args[1:]
	port, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalln(err)
	}
	p := peergossip.NewPeer(uint16(port))

	for {
		var peer int
		fmt.Println("Register peer port: ")
		fmt.Scan(&peer)
		p.Register(peer)
	}
}
