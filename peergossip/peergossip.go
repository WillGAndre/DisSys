package peergossip

import (
	"context"
	"fmt"
	"log"
	mrand "math/rand"
	"net"
	"strconv"
	"strings"
	"time"
	"token-ring/common"
	grpcapi "token-ring/grpcapi"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const K = 5
const SAMPLES = 25 // 100

type Peer struct {
	Port     uint16   `json:"port"`
	Registry []uint16 `json:"registry"`
	WordList []string `json:"wordlist"`
	Addr     net.IP   `json:"addr"`
	grpcapi.UnimplementedShellServer
}

func NewPeer(port uint16) *Peer {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	p := &Peer{
		Port:     port,
		Registry: make([]uint16, 0),
		WordList: make([]string, 0),
		Addr:     conn.LocalAddr().(*net.UDPAddr).IP,
	}
	go Listen(p)
	go p.PoissonWordProcess(SAMPLES)
	return p
}

func Listen(p *Peer) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", p.Port))
	if err != nil {
		log.Fatalln(err)
	}
	grpcs := grpc.NewServer()
	grpcapi.RegisterShellServer(grpcs, p)
	if err := grpcs.Serve(l); err != nil {
		log.Fatalln(err)
	}
}

// frequency of 1 event each 30 seconds
func (p *Peer) PoissonWordProcess(samples uint) {
	var ut float64
	timestamps := make([]float64, 0)
	i := 1

	// set seed
	mrand.Seed(int64(p.Port))

	// gen event timestamps, cast abstract time to sec (60s -- 1min ; λ = 2 ; 2 evs per 60s <=> 1ev per 30s)
	for i < int(samples) {
		ut += common.PoissonProcessTimeToNextEvent()
		timestamps = append(timestamps, ut*60)
		i += 1
	}

	// orchestrate events
	i = 0
	start := time.Now()
	for i < len(timestamps) {
		v := timestamps[i]
		if time.Since(start).Seconds() >= v {
			lines := common.GetLines("engmix.txt")
			wdi := mrand.Intn(len(lines))
			wd := lines[wdi]

			p.WordList = append(p.WordList, wd)
			go p.Gossip(wd, p.Port)
			i += 1
		} else {
			evtime := start.Add(time.Duration(v))
			time.Sleep(time.Since(evtime))
		}
	}
}

func (p *Peer) Register(regaddr int) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf(":%d", regaddr), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	g := grpcapi.NewShellClient(conn)
	res, err := g.Ping(context.Background(), &grpcapi.Message{Body: fmt.Sprintf("%d", p.Port)})
	if err != nil {
		log.Fatalf("error calling grpc call: %s\n", err)
	}

	pport, err := strconv.Atoi(res.Body)
	if err != nil {
		log.Fatalln(err)
	}
	p.Registry = append(p.Registry, uint16(pport))
	fmt.Printf("\t[%d] Ping Peer %d\n", p.Port, pport)
}

// gossipeer is the peer that "originaly" gossiped the word.
func (p *Peer) Gossip(word string, gossipeer uint16) {
	for _, peer := range p.Registry {
		if peer != gossipeer {
			fmt.Printf("\t[%d] Gossiping %s to %d\n", p.Port, word, peer)
			var conn *grpc.ClientConn
			conn, err := grpc.Dial(fmt.Sprintf(":%d", peer), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Fatalln(err)
			}
			defer conn.Close()

			g := grpcapi.NewShellClient(conn)
			_, err = g.Word(context.Background(), &grpcapi.Message{Body: fmt.Sprintf("%s:%d", word, p.Port)})
			if err != nil {
				log.Fatalf("error calling grpc call: %s\n", err)
			}
		}
	}
}

func (p *Peer) Ping(ctx context.Context, in *grpcapi.Message) (*grpcapi.Message, error) {
	addr, err := strconv.Atoi(in.Body)
	if err != nil {
		log.Fatalln(err)
	}

	p.Registry = append(p.Registry, uint16(addr))
	fmt.Printf("\t[%d] Ping from %d\n", p.Port, addr)
	return &grpcapi.Message{Body: fmt.Sprintf("%d", p.Port)}, nil
}

func (p *Peer) Word(ctx context.Context, in *grpcapi.Message) (*grpcapi.Message, error) {
	split := strings.Split(in.Body, ":")
	fmt.Printf("\t[%d] Received %s from %s\n", p.Port, split[0], split[1])
	word := split[0]
	pport, err := strconv.Atoi(split[1])
	if err != nil {
		log.Fatalln(err)
	}
	new := true
	for _, hword := range p.WordList {
		if word == hword {
			new = false
		}
	}

	if new {
		p.WordList = append(p.WordList, word)
		go p.Gossip(word, uint16(pport))
	} else {
		prob := mrand.Float64()
		if prob >= (1.0 / K) {
			p.WordList = append(p.WordList, word)
			go p.Gossip(word, uint16(pport))
		} else {
			fmt.Printf("\t[%d] Found repeated message '%s' from %s: Stopping gossiping\n", p.Port, split[0], split[1])
		}
	}

	return &grpcapi.Message{Body: ""}, nil
}
