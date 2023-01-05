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
// 	decremented for each token action (Bind)
const TTL = 4

type Peer struct {
	Port  uint16 `json:"port"`
	Next  uint16 `json:"next"`
	Token int    `json:"token"`
	Addr  net.IP `json:"addr"`
	TTL   int    `json:"ttl"`
	Lock  uint8  `json:"lock"`
	grpcapi.UnimplementedShellServer
}

func NewPeer(port uint16, next uint16, lock uint8) Peer {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	p := Peer{
		Port:  port,
		Next:  next,
		Token: 0,
		TTL:   TTL,
		Lock:  lock,
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
	res, err := g.Ping(context.Background(), &grpcapi.Message{Body: fmt.Sprintf("t:%d", p.Token)})
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

func (p *Peer) PeerShell() {
	for {
		fmt.Printf("commands: status, lock, unlock, fw, exit\n> ")
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		switch input.Text() {
		case "status":
			fmt.Printf("\n%+v\n", p)
		case "lock":
			LockPeer(int(p.Port), true)
		case "unlock":
			LockPeer(int(p.Port), false)
		case "fw":
			p.Bind()
			return
		case "exit":
			return
			// os.Exit(0)
		}
	}
}

// grpc request to lock peer by addr
func LockPeer(addr int, actionType bool) {
	var res *grpcapi.Message
	var conn *grpc.ClientConn
	conn, err := grpc.Dial(fmt.Sprintf(":%d", addr), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	g := grpcapi.NewShellClient(conn)
	if actionType {
		res, err = g.Ping(context.Background(), &grpcapi.Message{Body: "l:1"}) // lock
		if err != nil {
			log.Fatalf("error calling grpc call: %s\n", err)
		}
	} else {
		res, err = g.Ping(context.Background(), &grpcapi.Message{Body: "l:0"}) // unlock
		if err != nil {
			log.Fatalf("error calling grpc call: %s\n", err)
		}
	}
	log.Printf("\tPeer %d status: %s", addr, res)
}

// grpcapi calls

func (p *Peer) Ping(ctx context.Context, in *grpcapi.Message) (*grpcapi.Message, error) {
	msg := in.Body
	res := ""
	if strings.Contains(in.Body, "t") {
		msg = strings.Split(msg, ":")[1]
		token, err := strconv.Atoi(msg)
		if err != nil {
			log.Fatalln(err)
		}

		if p.TTL <= 0 {
			return &grpcapi.Message{Body: "1:1"}, nil
		}

		if p.Lock == 0 {
			fmt.Printf("\n%s\n", fmt.Sprintf("Token: %d\tPeer: %d", token, p.Port))
			p.Token = token + 1
			p.TTL -= 1
			p.Bind()
		} else if p.Lock == 1 {
			return &grpcapi.Message{Body: "2:2"}, nil
		}
		res = fmt.Sprintf("0:%d", token)
	} else if strings.Contains(in.Body, "l") {
		lock := strings.Split(msg, ":")[1]
		if lock == "1" {
			p.Lock = 1
		} else if lock == "0" {
			p.Lock = 0
		}
		res = fmt.Sprintf("%+v", p)
	}
	return &grpcapi.Message{Body: res}, nil

	// token, err := strconv.Atoi(in.Body)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// if p.TTL <= 0 {
	// 	return &grpcapi.Message{Body: "1:1"}, nil
	// }

	// if p.Lock == 0 {
	// 	fmt.Printf("\n%s\n", fmt.Sprintf("Token: %d\tPeer: %d", token, p.Port))
	// 	p.Token = token + 1
	// 	p.TTL -= 1
	// 	p.Bind()
	// } else if p.Lock == 1 {
	// 	return &grpcapi.Message{Body: "2:2"}, nil
	// }
	// return &grpcapi.Message{Body: fmt.Sprintf("0:%d", token)}, nil
}
