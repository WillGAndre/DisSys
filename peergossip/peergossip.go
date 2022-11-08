package peergossip

import (
	"context"
	"fmt"
	"log"
	mrand "math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
	"token-ring/common"
	grpcapi "token-ring/grpcapi"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const K = 5
const SAMPLES = 100

type Peer struct {
	Port     uint16            `json:"port"`
	Registry map[uint16]uint16 `json:"registry"`
	WordList []string          `json:"wordlist"`
	Count    uint16            `json:"count"` // <= K
	Addr     net.IP            `json:"addr"`
	mu       sync.Mutex
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
		Registry: make(map[uint16]uint16),
		WordList: make([]string, 0),
		Count:    0,
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

	// gen event timestamps
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

			p.mu.Lock()
			p.WordList = append(p.WordList, wd)
			p.mu.Unlock()
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
	p.mu.Lock()
	p.Registry[p.Count] = uint16(pport)
	p.Count++
	p.mu.Unlock()

	fmt.Printf("\t[%d] Ping Peer %d\n", p.Port, pport)
}

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

	p.mu.Lock()
	p.Registry[p.Count] = uint16(addr)
	p.Count++
	p.mu.Unlock()

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
		p.mu.Lock()
		p.WordList = append(p.WordList, word)
		p.mu.Unlock()
		go p.Gossip(word, uint16(pport))
	} else {
		prob := mrand.Float64()
		if prob >= 1/K {
			go p.Gossip(word, uint16(pport))
		}
	}

	return &grpcapi.Message{Body: ""}, nil
}
