package multicast

import (
	"context"
	"crypto/md5"
	b64 "encoding/base64"
	"fmt"
	"io"
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

const SAMPLES = 100

var VERBOSE = false
var STOP = false

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
	return p
}

// We are assuming 2 events per second, thus we multiply each event time by two
// 2 evs per 2 secs <=> 1 ev per sec (goal)
func (p *Peer) BootEvents() {
	var ut float64
	timestamps := make([]float64, 0)
	i := 1
	md := md5.New()
	// gen event timestamps
	for i < SAMPLES {
		ut += common.PoissonProcessTimeToNextEvent()
		timestamps = append(timestamps, ut*2)
		i += 1
	}
	// orchestrate events
	i = 0
	start := time.Now()
	if VERBOSE {
		fmt.Printf("\tEvents relative to [%d] @ %s: %f\n", p.Port, start, timestamps)
	}
	go func(p *Peer) {
		for i < len(timestamps) {
			v := timestamps[i]
			if time.Since(start).Seconds() >= v {
				io.WriteString(md, fmt.Sprintf("%f", v))
				fresh := fmt.Sprintf("%x", md.Sum(nil))[0:4]
				p.PingAll(fmt.Sprintf("ping-%s", fresh))
				i += 1
			} else {
				evtime := start.Add(time.Duration(v))
				time.Sleep(time.Since(evtime))
			}
		}
	}(p)
}

// Serves grpc
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

// ping peer: update clock and registry accordingly
func (p *Peer) PingPeer(addr int, msg string) string {
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
	appOut, _ := b64.StdEncoding.DecodeString(out[0])
	addr, err = strconv.Atoi(out[1])
	if err != nil {
		log.Fatalln(err)
	}
	clock, err := strconv.Atoi(out[2])
	if err != nil {
		log.Fatalln(err)
	}

	if !common.Contains(p.Registry, uint16(addr)) {
		p.Registry = append(p.Registry, uint16(addr))
	}
	if p.Clock <= uint16(clock) {
		p.Clock = uint16(clock) + 1
	}

	return string(appOut)
}

// grpc implementation of ping
func (p *Peer) Ping(ctx context.Context, in *grpcapi.Message) (*grpcapi.Message, error) {
	resOut := ""
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

	// add to queue
	if strings.Contains(msg[0], "ping") {
		p.Queue = p.OrderedInsert(in.Body, clock)
		p.PingAll(fmt.Sprintf("ack-%d-%d", addr, clock))
	} else if strings.Contains(msg[0], "ack") {
		ackout := strings.Split(msg[0], "-")
		pingIdx := p.AckCheck(ackout[1], ackout[2])
		if pingIdx != -1 {
			pingMsg := p.Queue[pingIdx]
			p.Queue = common.RemoveByIndex(p.Queue, pingIdx)
			if p.Gold {
				// fmt.Printf("\t[%d;%d] ACK: {%s} ; Queue: %s\n", p.Port, p.Clock, pingMsg, p.Queue)
				restr := fmt.Sprintf("\t[%d;%d] ACK {%s} ; Queue %s\n", p.Port, p.Clock, pingMsg, p.Queue)
				resOut = b64.StdEncoding.EncodeToString([]byte(restr))
			}
		}
		// else if VERBOSE {
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

			This is exactly what was implemented here, ACKs without correspondence can still be pushed to the queue, but shouldn't be printed.
		*/
	}

	return &grpcapi.Message{Body: fmt.Sprintf("%s:%d:%d", resOut, p.Port, p.Clock)}, nil
}

// multicast
func (p *Peer) PingAll(msg string) {
	logstr := ""
	if strings.Contains(msg, "ping") || VERBOSE {
		logstr = fmt.Sprintf("\t[%d] Multicast {%s} \n", p.Port, msg)
	}

	if strings.Contains(msg, "ping") {
		logstr += "\n\t------------------------------------\n\n"
	}
	for _, addr := range p.Registry {
		if VERBOSE {
			logstr += fmt.Sprintf("\t\t[%d;%d] ----> %d\n", p.Port, p.Clock, addr)
		}
		appOut := p.PingPeer(int(addr), msg)
		if appOut != "" {
			logstr += appOut
		}
	}
	if !STOP {
		fmt.Print(logstr)
	}
}

// Searches for index to insert msg (based on clock)
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

// Searches for a "ping" msg that was issued by the ack addr and at the ack time (clock)
func (p *Peer) AckCheck(ackaddr string, ackclock string) int {
	for i, v := range p.Queue {
		split := strings.Split(v, ":")
		if strings.Contains(split[0], "ping") {
			vackclock, err := strconv.Atoi(ackclock)
			if err != nil {
				log.Fatalln(err)
			}
			vclock, err := strconv.Atoi(split[2])
			if err != nil {
				log.Fatalln(err)
			}
			if ackaddr == split[1] && vackclock == vclock {
				return i
			}
		}
	}
	return -1
}
