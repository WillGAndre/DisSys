package multicast

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
	common "token-ring/common"
	grpcapi "token-ring/grpcapi"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const MAX_PEERS = 8
const MAX_MSGS = 25
const SAMPLES = 100

/** Gold will only print received messages (order messages in queue by timestamps - Totally Ordered Multicast)
 * \- After receiving msg, return the global queue of msgs
**/
type Gold struct {
	Port     uint16   `json:"port"`
	Remote   uint16   `json:"remote"`
	Registry []uint16 `json:"registry"`
	Queue    []string `json:"queue"`
	Addr     net.IP   `json:"addr"`
	grpcapi.UnimplementedShellServer
}

/** Peer will only send msgs to Gold peers (msgs must be timestamped with Lamport Clock - counter) **/
type Peer struct {
	Port     uint16   `json:"port"`
	Registry []uint16 `json:"registry"`
	Storage  []uint16 `json:"storage"`
	Clock    uint16   `json:"clock"`
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
		Clock:    0,
		Registry: make([]uint16, 0),
		Storage:  make([]uint16, 0),
		Addr:     conn.LocalAddr().(*net.UDPAddr).IP,
	}
	go Listen(p)
	go p.MulticastProcess(SAMPLES)
	return p
}

func NewGoldPeer(port uint16, remote uint16) *Gold {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	p := &Gold{
		Port:     port,
		Remote:   remote,
		Registry: make([]uint16, 0),
		Queue:    make([]string, 0),
		Addr:     conn.LocalAddr().(*net.UDPAddr).IP,
	}
	go Listen(p)
	go func(p *Gold) {
		for {
			time.Sleep(10 * time.Second)
			p.ExchangeQueue(int(p.Remote))
		}
	}(p)
	return p
}

func Listen(p interface{}) {
	switch p := p.(type) {
	case *Peer:
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", p.Port))
		if err != nil {
			log.Fatalln(err)
		}
		grpcs := grpc.NewServer()
		grpcapi.RegisterShellServer(grpcs, p)
		if err := grpcs.Serve(l); err != nil {
			log.Fatalln(err)
		}
	case *Gold:
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
}

func (p *Peer) MulticastProcess(samples uint) {
	var ut float64
	timestamps := make([]float64, 0)
	i := 1

	// gen event timestamps
	for i < int(samples) {
		ut += common.PoissonProcessTimeToNextEvent()
		timestamps = append(timestamps, ut*60)
		i += 1
	}

	// orchestrate events
	i = 0
	start := time.Now()
	fmt.Printf("\tEvents relative to [%d] @ %s: %f\n", p.Port, start, timestamps)
	for i < len(timestamps) {
		v := timestamps[i]
		if time.Since(start).Seconds() >= v {
			go p.PingAll()
			i += 1
		} else {
			evtime := start.Add(time.Duration(v))
			time.Sleep(time.Since(evtime))
		}
	}
}

func (p *Peer) PingPeer(addr int) {
	p.Clock += 1
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf(":%d", addr), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	g := grpcapi.NewShellClient(conn)
	msg := fmt.Sprintf("%d:%d", p.Port, p.Clock)
	fmt.Printf("\t[%d;%d] Ping Peer %d\n", p.Port, p.Clock, addr)
	res, err := g.Ping(context.Background(), &grpcapi.Message{Body: msg})
	if err != nil {
		log.Fatalf("error calling grpc call: %s\n", err)
	}

	out := strings.Split(res.Body, ":")
	addr, err = strconv.Atoi(out[0])
	if err != nil {
		log.Fatalln(err)
	}
	clock, err := strconv.Atoi(out[1])
	if err != nil {
		log.Fatalln(err)
	}

	if !common.Contains(p.Registry, uint16(addr)) {
		p.Registry = append(p.Registry, uint16(addr))
	}
	if p.Clock <= uint16(clock) {
		p.Clock = uint16(clock) + 1
	}
	// fmt.Printf("\t[%d;%d] ACK Ping from %d\n\n", p.Port, p.Clock, addr)
}

func (p *Peer) PingGold(addr int) {
	p.Clock += 1
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf(":%d", addr), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	g := grpcapi.NewShellClient(conn)
	msg := fmt.Sprintf("%d:%d", p.Port, p.Clock)
	fmt.Printf("\t[%d;%d] Ping Gold Peer %d\n", p.Port, p.Clock, addr)
	_, err = g.Ping(context.Background(), &grpcapi.Message{Body: msg})
	if err != nil {
		log.Fatalf("error calling grpc call: %s\n", err)
	}

	// out := res.Body
	if !common.Contains(p.Storage, uint16(addr)) {
		p.Storage = append(p.Storage, uint16(addr))
	}

	// fmt.Printf("\t[%d;%d] ACK Gold Peer queue: %s\n\n", p.Port, p.Clock, out)
}

func (p *Peer) PingGolds() {
	if len(p.Storage) == 0 {
		return
	}
	for _, addr := range p.Storage {
		go p.PingGold(int(addr))
	}
}

func (p *Peer) PingAll() {
	for _, addr := range p.Registry {
		go p.PingPeer(int(addr))
	}
	p.PingGolds()
}

/** Peer impl of Ping **/
func (p *Peer) Ping(ctx context.Context, in *grpcapi.Message) (*grpcapi.Message, error) {
	msg := strings.Split(in.Body, ":")
	addr, err := strconv.Atoi(msg[0])
	if err != nil {
		log.Fatalln(err)
	}
	clock, err := strconv.Atoi(msg[1])
	if err != nil {
		log.Fatalln(err)
	}

	if !common.Contains(p.Registry, uint16(addr)) {
		p.Registry = append(p.Registry, uint16(addr))
	}
	if p.Clock <= uint16(clock) {
		p.Clock = uint16(clock) + 1
	}

	fmt.Printf("\t[%d;%d] ACK Ping from %d\n\n", p.Port, p.Clock, addr)
	return &grpcapi.Message{Body: fmt.Sprintf("%d:%d", p.Port, p.Clock)}, nil
}

/** Gold impl of Ping **/
func (g *Gold) Ping(ctx context.Context, in *grpcapi.Message) (*grpcapi.Message, error) {
	msg := in.Body
	queue := fmt.Sprintf("%s", g.Queue)

	if strings.Contains(msg, "[") {
		qlog := strings.Split(queue, " ")
		mlog := strings.Split(msg, " ")

		if len(qlog) == len(mlog) {
			for i := range qlog {
				if strings.Split(qlog[i], ":")[0] != strings.Split(mlog[i], ":")[0] {
					goto err
				}
			}
			fmt.Printf("\t[!;%d] Gold Peer Synchronized with [%d]\n\n", g.Port, g.Remote)
		} else {
			goto err
		}
	} else if strings.Contains(msg, ":") {
		out := strings.Split(msg, ":")
		addr, err := strconv.Atoi(out[0])
		if err != nil {
			log.Fatalln(err)
		}
		clock, err := strconv.Atoi(out[1])
		if err != nil {
			log.Fatalln(err)
		}

		if !common.Contains(g.Registry, uint16(addr)) {
			g.Registry = append(g.Registry, uint16(addr))
		}
		hit := false
		if len(g.Queue) > 1 {
			for i, v := range g.Queue {
				vclock, err := strconv.Atoi(strings.Split(v, ":")[1])
				if err != nil {
					log.Fatalln(err)
				}
				if clock > vclock {
					g.Queue = common.Insert(g.Queue, i, msg)
					hit = true
					break
				}
			}
		}
		if !hit {
			g.Queue = append(g.Queue, msg)
		}
		fmt.Printf("\t[%d] ACK Ping from %d; Current queue: %s\n\n", g.Port, addr, g.Queue)
	}
	goto res
err:
	fmt.Printf("\t[*;%d] Queue not synchronized ; Local queue: %s ; Remote queue: %s\n\n", g.Port, g.Queue, msg)
	return &grpcapi.Message{Body: "err: Queue not synchronized"}, nil

res:
	return &grpcapi.Message{Body: queue}, nil
}

func (g *Gold) ExchangeQueue(addr int) {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf(":%d", addr), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	gr := grpcapi.NewShellClient(conn)
	msg := fmt.Sprintf("%s", g.Queue)
	//fmt.Printf("\t[%d;%d] Ping Gold Peer %d\n", p.Port, p.Clock, addr)
	res, err := gr.Ping(context.Background(), &grpcapi.Message{Body: msg})
	if err != nil {
		log.Fatalf("error calling grpc call: %s\n", err)
	}

	if res.Body == "err: Queue not synchronized" {
		go g.ExchangeQueue(int(g.Remote))
	}

}

/*
	TODO:																														  Done ?
		- Add RPC call for communication between normal and Gold peers
			1) Peer --> Gold																										•
			2) Peer --> Peer 																										•
			3-extra) Gold --> Gold
				\
				 --> Share queue and compare, msg content (exclude clock) should be equal (ordered)

		- Messages between Peers must be timestamped with counter (clock), clock must be updated accordingly						•
		when a message is received (if local clock < received clock: local clock = received clock + 1)

		- When a Gold receives a msg, he must arrange his local queue according to the timestamps and then print it.				•
		If all goes accordingly multiple Golds should have the same queue sequence (Totally Ordered Multicast).

	Note:
		- Remove mutex locks from peergossip (?)
*/
