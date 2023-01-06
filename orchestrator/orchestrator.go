package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
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
	} else if poolType == 1 { // Peergossip module: assumed network topology
		for i := 0; i < size; i++ {
			addr, err := strconv.Atoi(fmt.Sprintf("%d%d", peerGossipPrefix, i))
			if err != nil {
				log.Fatalln(err)
			}
			pool = append(pool, peergossip.NewPeer(uint16(addr)))
		}
	} else if poolType == 2 { // Multicast module: assumed network topology
		i := 0
		for ; i < size; i++ {
			addr, err := strconv.Atoi(fmt.Sprintf("%d%d", peerMulticastPrefix, i))
			if err != nil {
				log.Fatalln(err)
			}
			pool = append(pool, multicast.NewPeer(uint16(addr), false))
		}
		j := i
		for ; j < (i + size/2); j++ {
			addr, err := strconv.Atoi(fmt.Sprintf("%d%d", peerMulticastPrefix, j))
			if err != nil {
				log.Fatalln(err)
			}
			pool = append(pool, multicast.NewPeer(uint16(addr), true))
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
		PoolPeer()
		Clear()
		PoolPeerGossip()
		Clear()
		PoolPeerMulticast()
	} else if *peerFlg {
		PoolPeer()
		Clear()
	} else if *gossipFlg {
		PoolPeerGossip()
		Clear()
	} else if *multicastFlg {
		PoolPeerMulticast()
		Clear()
	}
}

func PoolPeer() {
	fmt.Printf("%v %v\n", color.GreenString("Info: "), "Starting Peer Token Ring Module")
	fmt.Printf("%v %v", color.YellowString("Prompt: "), "Pool size (min: 4) > ")
	input := bufio.NewScanner(os.Stdin)
	input.Scan()
	size, err := strconv.Atoi(input.Text())
	if err != nil {
		log.Fatalln(err)
	}
	pool := initPeerPool(0, size)
	fmt.Printf("%v %v\n", color.GreenString("Info: "), fmt.Sprintf("Created %d peers with TTL = %d", len(pool), peer.TTL))
	fmt.Printf("%v %v\n", color.GreenString("Info: "), "shell - pop a shell in selected peer (useful for forwarding - fw and setting lock/unlock)")
	fmt.Printf("\n------------------------------------\n\n")

	for {
		fmt.Printf("%v commands: shell, reset, exit\n> ", color.CyanString("[Peer Token Ring Module Shell]"))
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
	fmt.Printf("%v %v\n", color.GreenString("Info: "), "start - start gossiping ; gossip - send custom gossip (after start)")
	fmt.Printf("\n------------------------------------\n\n")

	// topology from assignment - hardcoded
	p1, _ := pool[0].(*peergossip.Peer)
	p2, _ := pool[1].(*peergossip.Peer)
	p3, _ := pool[2].(*peergossip.Peer)
	p4, _ := pool[3].(*peergossip.Peer)
	p5, _ := pool[4].(*peergossip.Peer)
	p6, _ := pool[5].(*peergossip.Peer)
	// p2.Register(int(p1.Port))
	// p2.Register(int(p3.Port))
	// p2.Register(int(p4.Port))
	// p4.Register(int(p5.Port))
	// p4.Register(int(p6.Port))
	// ----

	for {
		fmt.Printf("%v commands: start, gossip, status, exit\n> ", color.CyanString("[Peer Gossip Module Shell]"))
		input.Scan()
		switch input.Text() {
		case "start":
			fmt.Printf("%v %v\n", color.GreenString("Info: "), "Gossip timestamps will be printed, exit will stop pool.")
			p2.Register(int(p1.Port))
			p2.Register(int(p3.Port))
			p2.Register(int(p4.Port))
			p4.Register(int(p5.Port))
			p4.Register(int(p6.Port))
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
		case "reset":
			for _, p := range pool {
				p.(*peergossip.Peer).Registry = make([]uint16, 0)
			}
			pool = make([]interface{}, 0)
			peerGossipPrefix += 1
			pool = initPeerPool(1, peergossip.K+1)
		case "exit":
			for _, p := range pool {
				p.(*peergossip.Peer).Registry = make([]uint16, 0)
			}
			return
		}
	}
}

func PoolPeerMulticast() {
	fmt.Printf("%v %v\n", color.GreenString("Info: "), "Starting Peer Multicast Module")
	pool := initPeerPool(2, 4)
	input := bufio.NewScanner(os.Stdin)
	fmt.Printf("%v %v\n", color.GreenString("Info: "), fmt.Sprintf("Created %d peers each with %d samples", 6, multicast.SAMPLES))
	fmt.Printf("%v Poisson Process: λ = 2 - 2evs per 2s - 1ev per second\n", color.GreenString("Info: "))
	fmt.Printf("%v %v\n", color.GreenString("Info: "), "start - start multicast (with optional verbose mode)")

	fmt.Printf("\n%v %v\n", color.GreenString("Info: "), "To stop pool send an interrupt, shell module will be displayed again.")
	fmt.Printf("------------------------------------\n\n")

	// topology from assignment - hardcoded
	p1, _ := pool[0].(*multicast.Peer)
	p2, _ := pool[1].(*multicast.Peer)
	p3, _ := pool[2].(*multicast.Peer)
	p4, _ := pool[3].(*multicast.Peer)
	p5, _ := pool[4].(*multicast.Peer)
	p6, _ := pool[5].(*multicast.Peer)

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
	// ----

	for {
		fmt.Printf("%v commands: start, reset, exit\n> ", color.CyanString("[Peer Multicast Module Shell]"))
		input.Scan()
		switch input.Text() {
		case "start":
			sigc := make(chan os.Signal, 1)
			signal.Notify(sigc,
				syscall.SIGHUP,
				syscall.SIGINT,
				syscall.SIGTERM,
				syscall.SIGQUIT)
			fmt.Printf("%v Verbose ? (y) ", color.YellowString("Prompt:"))
			input.Scan()
			opt := input.Text()
			if opt == "y" {
				multicast.VERBOSE = true
			}
			for _, p := range pool {
				p.(*multicast.Peer).BootEvents()
			}
			go func() {
				<-sigc
				multicast.STOP = true
				for _, p := range pool {
					p.(*multicast.Peer).Registry = make([]uint16, 0)
				}
				pool = make([]interface{}, 0)
			}()
			for {
				if len(pool) == 0 {
					break
				}
			}
			fmt.Printf("%v %v\n", color.RedString("Warn: "), "Peer Multicast Module must be reset (unstable - old prints might appear)")
		case "reset":
			multicast.STOP = false
			pool = make([]interface{}, 0)
			peerMulticastPrefix += 1
			pool = initPeerPool(2, 4)
		case "exit":
			return
		}
	}
}

// -- aux
// Ref: https://stackoverflow.com/a/22896706
var clear map[string]func()

func init() {
	clear = make(map[string]func())
	clear["linux"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["darwin"] = func() {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}
func Clear() {
	value, ok := clear[runtime.GOOS]
	if ok {
		value()
	}
}
