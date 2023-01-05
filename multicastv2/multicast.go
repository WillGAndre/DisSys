package multicastv2

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
	"token-ring/common"
	grpcapi "token-ring/grpcapi"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const MAX_PEERS = 8
const MAX_MSGS = 25
const SAMPLES = 100

type Peer struct {
	Port     uint16   `json:"port"`
	Registry []uint16 `json:"registry"`
	Queue    []string `json:"queue"`
	Clock    uint16   `json:"clock"`
	Addr     net.IP   `json:"addr"`
	Gold     bool     `json:"gold"`
	grpcapi.UnimplementedShellServer
}

func NewPeer(port uint16, gold bool) *Peer {
	var ut float64
	timestamps := make([]float64, 0)
	i := 1
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	p := &Peer{
		Port:     port,
		Clock:    0,
		Registry: make([]uint16, 0),
		Addr:     conn.LocalAddr().(*net.UDPAddr).IP,
		Gold:     gold,
	}
	go Listen(p)
	// gen event timestamps
	for i < SAMPLES {
		ut += common.PoissonProcessTimeToNextEvent()
		timestamps = append(timestamps, ut*2)
		i += 1
	}
	// orchestrate events
	i = 0
	start := time.Now()
	fmt.Printf("\tEvents relative to [%d] @ %s: %f\n", p.Port, start, timestamps)
	go func(p *Peer) {
		for i < len(timestamps) {
			v := timestamps[i]
			if time.Since(start).Seconds() >= v {
				p.PingAll("ping")
				i += 1
			} else {
				evtime := start.Add(time.Duration(v))
				time.Sleep(time.Since(evtime))
			}
		}
	}(p)
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

func (p *Peer) PingPeer(addr int, msg string) {
	p.Clock += 1
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf(":%d", addr), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	g := grpcapi.NewShellClient(conn)
	res, err := g.Ping(context.Background(), &grpcapi.Message{Body: fmt.Sprintf("%s:%d:%d", msg, p.Port, p.Clock)})
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
}

func (p *Peer) Ping(ctx context.Context, in *grpcapi.Message) (*grpcapi.Message, error) {
	msg := strings.Split(in.Body, ":")
	addr, err := strconv.Atoi(msg[1])
	if err != nil {
		log.Fatalln(err)
	}
	clock, err := strconv.Atoi(msg[2])
	if err != nil {
		log.Fatalln(err)
	}

	if !common.Contains(p.Registry, uint16(addr)) {
		p.Registry = append(p.Registry, uint16(addr))
	}
	if p.Clock <= uint16(clock) {
		p.Clock = uint16(clock) + 1
	}

	if msg[0] == "ping" {
		p.Queue = p.OrderedInsert(in.Body, clock)
		p.PingAll(fmt.Sprintf("ack-%d-%d", addr, clock))
		// fmt.Printf("\t[%d;%d] QUEUE = %s\n", p.Port, p.Clock, p.Queue)
	} else if strings.Contains(msg[0], "ack") {
		ackout := strings.Split(msg[0], "-")
		pingIdx := p.AckCheck(ackout[1], ackout[2])
		if pingIdx != -1 {
			pingMsg := p.Queue[pingIdx]
			p.Queue = common.RemoveByIndex(p.Queue, pingIdx)
			fmt.Printf("\t[%d;%d] ACK: {%s} ; Queue: %s\n", p.Port, p.Clock, pingMsg, p.Queue)
			// Print ACK if gold (without acks)
		}
		// set as verbose
		// else {
		// 	p.Queue = p.OrderedInsert(in.Body, clock)
		// }
		/*
			As per the pin in Slack:
			_____________________________________________________________________________________________________________________________
				No exercício 3, quando se diz que um “ack” que esteja no início da fila deve ser descartado,
				isso quer dizer que não é passado à aplicação (neste caso) para ser imprimido. No entanto,
				para tirarem o ack da fila a condição a cumprir é a mesma para com todas as mensagens normais:
				só podem tirar o ack da cabeça da fila se no resto da mesma tiverem mensagens de todos os outros processos.
			_____________________________________________________________________________________________________________________________

			This is exactly what was implemented here, ACKs without correspondence are still pushed to the queue, but aren't printed (at
			the application). For verbose, remove cmts in else above.
		*/
	}

	return &grpcapi.Message{Body: fmt.Sprintf("%d:%d", p.Port, p.Clock)}, nil
}

func (p *Peer) PingAll(msg string) {
	// fmt.Printf("\t[%d;%d] Multicast {%s} \n", p.Port, p.Clock, msg)
	logstr := fmt.Sprintf("\t[%d] Multicast {%s} \n", p.Port, msg)
	for _, addr := range p.Registry {
		logstr += fmt.Sprintf("\t\t[%d;%d] ----> %d\n", p.Port, p.Clock, addr)
		p.PingPeer(int(addr), msg)
	}
	fmt.Print(logstr)
}

func (p *Peer) OrderedInsert(msg string, clock int) []string {
	hit := false
	for i, v := range p.Queue {
		vclock, err := strconv.Atoi(strings.Split(v, ":")[2])
		if err != nil {
			log.Fatalln(err)
		}
		if clock > vclock {
			p.Queue = common.Insert(p.Queue, i, msg)
			hit = true
			break
		}
	}
	if !hit {
		p.Queue = append(p.Queue, msg)
	}
	return p.Queue
}

func (p *Peer) AckCheck(ackaddr string, ackclock string) int {
	for i, v := range p.Queue {
		split := strings.Split(v, ":")
		if split[0] == "ping" {
			vackclock, err := strconv.Atoi(ackclock)
			if err != nil {
				log.Fatalln(err)
			}
			vclock, err := strconv.Atoi(split[2])
			if err != nil {
				log.Fatalln(err)
			}
			if ackaddr == split[1] && vackclock == vclock { // >= vclock
				return i
			}
		}
	}
	return -1
}
