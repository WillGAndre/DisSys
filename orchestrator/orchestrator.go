package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"token-ring/multicast"
	"token-ring/peer"
	"token-ring/peergossip"

	"github.com/fatih/color"
)

type Pool []interface{}

var peerPrefix = 444
var peerGossipPrefix = 464
var peerMulticastPrefix = 474

func initPeerPool(poolType int, size int) Pool {
	pool := Pool{}
	if poolType == 0 { // Peer Module
		if size < 4 {
			size = 4
		}
		for i := 0; i < size; i++ {
			addr, err := strconv.Atoi(fmt.Sprintf("%d%d", peerPrefix, i))
			if err != nil {
				log.Fatalln(err)
			}
			next, err := strconv.Atoi(fmt.Sprintf("%d%d", peerPrefix, i+1))
			if err != nil {
				log.Fatalln(err)
			}
			if i == size-1 {
				next, err = strconv.Atoi(fmt.Sprintf("%d0", peerPrefix))
				if err != nil {
					log.Fatalln(err)
				}
			}
			pool = append(pool, peer.NewPeer(uint16(addr), uint16(next), 0))
		}
	} else if poolType == 1 { // Peergossip module: assumed fixed size
		for i := 0; i < size; i++ {
			addr, err := strconv.Atoi(fmt.Sprintf("%d%d", peerGossipPrefix, i))
			if err != nil {
				log.Fatalln(err)
			}
			pool = append(pool, peergossip.NewPeer(uint16(addr)))
		}
	} else if poolType == 2 {
		i := 0
		for ; i < size; i++ {
			addr, err := strconv.Atoi(fmt.Sprintf("%d%d", peerMulticastPrefix, i))
			if err != nil {
				log.Fatalln(err)
			}
			pool = append(pool, multicast.NewPeer(uint16(addr)))
		}
		j := i
		for ; i < (i + size/2); i++ {
			addr, err := strconv.Atoi(fmt.Sprintf("%d%d", peerMulticastPrefix, i))
			if err != nil {
				log.Fatalln(err)
			}
			if i == ((i + size/2) - 1) {
				fst, err := strconv.Atoi(fmt.Sprintf("%d%d", peerMulticastPrefix, j))
				if err != nil {
					log.Fatalln(err)
				}
				pool = append(pool, multicast.NewGoldPeer(uint16(addr), uint16(fst)))
			} else {
				nxt, err := strconv.Atoi(fmt.Sprintf("%d%d", peerMulticastPrefix, i+1))
				if err != nil {
					log.Fatalln(err)
				}
				pool = append(pool, multicast.NewGoldPeer(uint16(addr), uint16(nxt)))
			}
		}
	}
	return pool
}

func main() {
	peerFlg := flag.Bool("peer", false, "run peer module")
	gossipFlg := flag.Bool("peer-gossip", false, "run peer-gossip module")
	multicastFlg := flag.Bool("multicast", false, "run multicast module")
	flag.Parse()

	if !*peerFlg && !*gossipFlg && !*multicastFlg {
		// PoolPeer()
		// PoolPeerGossip()
		PoolPeerMulticast()
		// exec all
	}
}

func PoolPeer() {
	fmt.Printf("%v %v\n", color.GreenString("Info: "), "Starting Peer Module")
	fmt.Printf("%v %v", color.YellowString("Prompt: "), "Pool size (min: 4) > ")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	size, err := strconv.Atoi(input.Text())
	if err != nil {
		log.Fatalln(err)
	}
	pool := initPeerPool(0, size)
	fmt.Printf("%v %v\n", color.GreenString("Info: "), fmt.Sprintf("Created %d peers with TTL = %d", len(pool), peer.TTL))
	fmt.Printf("\n------------------------------------\n\n")

	for {
		fmt.Printf("%v commands: shell, reset, exit\n> ", color.CyanString("[Peer Module Shell]"))
		input.Scan()
		switch input.Text() {
		case "shell":
			fmt.Printf("%v %v\n", color.GreenString("Info: "), fmt.Sprintf("Peers: 0...%d", len(pool)-1))
			fmt.Printf("idx > ")
			input.Scan()
			p, err := strconv.Atoi(input.Text())
			if err != nil {
				log.Fatalln(err)
			}
			if p >= 0 && p < len(pool) {
				pr, ok := pool[p].(peer.Peer)
				if ok {
					pr.PeerShell()
				}
			}
		case "reset":
			pool = make([]interface{}, 0)
			peerPrefix += 1
			pool = initPeerPool(0, size)
		case "exit":
			return
		}
	}
}

func PoolPeerGossip() {
	fmt.Printf("%v %v\n", color.GreenString("Info: "), "Starting Peer Gossip Module")
	pool := initPeerPool(1, peergossip.K+1)
	input := bufio.NewScanner(os.Stdin)
	fmt.Printf("%v %v\n", color.GreenString("Info: "), fmt.Sprintf("Created %d peers each with %d samples", peergossip.K+1, peergossip.SAMPLES))
	fmt.Printf("%v Poisson Process: λ = 2 - 2evs per 60s - 1ev per 30s\n", color.GreenString("Info: "))
	fmt.Printf("\n------------------------------------\n\n")

	p1, _ := pool[0].(*peergossip.Peer)
	p2, _ := pool[1].(*peergossip.Peer)
	p3, _ := pool[2].(*peergossip.Peer)
	p4, _ := pool[3].(*peergossip.Peer)
	p5, _ := pool[4].(*peergossip.Peer)
	p6, _ := pool[5].(*peergossip.Peer)

	p2.Register(int(p1.Port))
	p2.Register(int(p3.Port))
	p2.Register(int(p4.Port))
	p4.Register(int(p5.Port))
	p4.Register(int(p6.Port))

	for {
		fmt.Printf("%v commands: gossip, status, exit\n> ", color.CyanString("[Peer Gossip Module Shell]"))
		input.Scan()
		switch input.Text() {
		case "gossip":
			fmt.Printf("%v %v\n", color.GreenString("Info: "), fmt.Sprintf("Peers: 0...%d", len(pool)-1))
			fmt.Printf("Sender idx > ")
			input.Scan()
			p, err := strconv.Atoi(input.Text())
			if err != nil {
				log.Fatalln(err)
			}

			if p >= 0 && p < len(pool) {
				sd, _ := pool[p].(*peergossip.Peer)
				fmt.Printf("%v", color.YellowString("Prompt: "))
				input.Scan()
				msg := input.Text()
				sd.Gossip(msg, sd.Port)
			}
		case "status":
			for i, p := range pool {
				fmt.Printf("[%d] %+v\n", i, p.(*peergossip.Peer))
			}
		case "exit":
			return
		}
	}
}

func PoolPeerMulticast() {
	// pool := initPeerPool(1, 4)
	// TODO: finish testing multicast EVERY MSG MUST BE MULTICAST - including ACKS
	// - go test -test.run TestMulticastProc
}

func HandleSig(exit bool) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		// clear cluster
		if exit {
			os.Exit(0)
		}
	}()
}
