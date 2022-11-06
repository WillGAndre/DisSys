package peergossip

import (
	"context"
	"fmt"
	"log"
	mrand "math/rand"
	"net"
	"strconv"
	"sync"
	"time"
	"token-ring/common"
	grpcapi "token-ring/grpcapi"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const K = 5

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
	go p.PoissonWordProcess()
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
func (p *Peer) PoissonWordProcess() {
	for { // literal translation
		lines := common.GetLines()
		wdi := mrand.Intn(len(lines))
		wd := lines[wdi]

		p.mu.Lock()
		p.WordList = append(p.WordList, wd)
		p.mu.Unlock()
		go p.Gossip(wd)
		time.Sleep(30 * time.Second)
	}

	// for {
	// 	events := common.PoissonProcessEvents(1)
	// 	for events > 0 {
	// 		lines := common.GetLines()
	// 		wdi := mrand.Intn(len(lines))
	// 		wd := lines[wdi]
	// 		p.mu.Lock()
	// 		p.WordList = append(p.WordList, wd)
	// 		p.mu.Unlock()
	// 		go p.Gossip(wd)
	// 		events -= 1
	// 	}
	// 	time.Sleep(time.Duration(common.PoissonProcessTimeToNextEvent()) * time.Second)
	// }
}

func (p *Peer) Register(regaddr int) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf(":%d", regaddr), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	// log.Printf("%s\n", fmt.Sprintf("[%d] -> [%d] Token: %d", p.Port, p.Next, p.Token))
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
}

func (p *Peer) Gossip(word string) {
	for _, peer := range p.Registry {
		var conn *grpc.ClientConn
		conn, err := grpc.Dial(fmt.Sprintf(":%d", peer), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalln(err)
		}
		defer conn.Close()

		g := grpcapi.NewShellClient(conn)
		_, err = g.Word(context.Background(), &grpcapi.Message{Body: word})
		if err != nil {
			log.Fatalf("error calling grpc call: %s\n", err)
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
	return &grpcapi.Message{Body: fmt.Sprintf("%d", p.Port)}, nil
}

func (p *Peer) Word(ctx context.Context, in *grpcapi.Message) (*grpcapi.Message, error) {
	word := in.Body
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
		go p.Gossip(word)
	} else {
		// TODO: stop gossip with prob 1/k
	}

	return &grpcapi.Message{Body: ""}, nil
}
