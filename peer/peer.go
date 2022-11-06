package peer

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	grpcapi "token-ring/grpcapi"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Peer TTL
const TTL = 4

// Peer Lock
var plock = 0

type Peer struct {
	Port  uint16 `json:"port"`
	Next  uint16 `json:"next"`
	Token int    `json:"token"`
	Addr  net.IP `json:"addr"`
	TTL   int    `json:"ttl"`
	grpcapi.UnimplementedShellServer
}

func NewPeer(port uint16, next uint16, lock uint8) Peer {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()
	plock = int(lock)

	p := Peer{
		Port:  port,
		Next:  next,
		Token: 0,
		TTL:   TTL,
		Addr:  conn.LocalAddr().(*net.UDPAddr).IP,
	}
	go Listen(p)
	return p
}

func Listen(p Peer) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", p.Port))
	if err != nil {
		log.Fatalln(err)
	}
	grpcs := grpc.NewServer()
	grpcapi.RegisterShellServer(grpcs, &p)
	if err := grpcs.Serve(l); err != nil {
		log.Fatalln(err)
	}
}

func (p *Peer) Bind() {
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf(":%d", p.Next), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	// log.Printf("%s\n", fmt.Sprintf("[%d] -> [%d] Token: %d", p.Port, p.Next, p.Token))
	g := grpcapi.NewShellClient(conn)
	res, err := g.Ping(context.Background(), &grpcapi.Message{Body: fmt.Sprintf("%d", p.Token)})
	if err != nil {
		log.Fatalf("error calling grpc call: %s\n", err)
	}

	rstr := strings.Split(res.Body, ":")
	rval, err := strconv.Atoi(rstr[0])
	if err != nil {
		log.Fatalln(err)
	}
	if rval == 1 {
		log.Printf("\tPeer %d TTL expired\n", p.Next)
	} else if rval == 2 {
		log.Printf("\tPeer %d locked\n", p.Next)
	}
}

func PeerShell(p Peer) {
	fmt.Println("cmds: lock, unlock, fw")
	for {
		fmt.Printf("> ")
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		switch input.Text() {
		case "status":
			fmt.Printf("\n%+v\n", p)
		case "lock":
			plock = 1
		case "unlock":
			plock = 0
		case "fw":
			p.Bind()
		case "exit":
			os.Exit(0)
		}
	}
}

// grpcapi calls

func (p *Peer) Ping(ctx context.Context, in *grpcapi.Message) (*grpcapi.Message, error) {
	token, err := strconv.Atoi(in.Body)
	if err != nil {
		log.Fatalln(err)
	}

	if p.TTL <= 0 {
		return &grpcapi.Message{Body: "1:1"}, nil
	}

	if plock == 0 {
		fmt.Printf("\n%s\n", fmt.Sprintf("Token: %d\tPeer: %d", token, p.Port))
		p.Token = token + 1
		p.TTL -= 1
		p.Bind()
	} else if plock == 1 {
		return &grpcapi.Message{Body: "2:2"}, nil
	}
	return &grpcapi.Message{Body: fmt.Sprintf("0:%d", token)}, nil
}
