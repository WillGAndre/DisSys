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
	"token-ring/peer"

	"github.com/fatih/color"
)

type Pool []interface{}

var peerPrefix = 444

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
				next = 4440
			}
			pool = append(pool, peer.NewPeer(uint16(addr), uint16(next), 0))
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
