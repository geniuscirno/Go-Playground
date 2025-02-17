package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/geniuscirno/Go-Playground/grpc/bench/stream/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	msgs = flag.Int("msgs", 100000, "the message to send")
	size = flag.Int("size", 128, "the size of each message")
	pub  = flag.Int("pub", 1, "number of concurrent publishers")
)

type Sample struct {
	JobMsgCnt uint64
	MsgCnt    uint64
	MsgBytes  uint64
	Start     time.Time
	End       time.Time
}

func (s *Sample) Throughput() float64 {
	return float64(s.MsgBytes) / s.Duration().Seconds()
}

func (s *Sample) Rate() int64 {
	return int64(float64(s.JobMsgCnt) / s.Duration().Seconds())
}

func (s *Sample) Duration() time.Duration {
	return s.End.Sub(s.Start)
}

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

func startServer(addr string) (func(), error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	done := make(chan struct{})
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
		done <- struct{}{}
	}()
	return func() {
		time.Sleep(1 * time.Second)
		log.Printf("server shutdown at %v\n", lis.Addr())
		s.GracefulStop()
		<-done
	}, nil
}

func (s *server) SayHello(stream pb.Greeter_SayHelloServer) error {
	var i int
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			log.Printf("Server received %d messages", i)
			return stream.SendAndClose(&pb.HelloReply{})
		}
		if err != nil {
			return err
		}
		i++
	}
}

type Client struct {
	Sample
	numMsg  int
	msgSize int
}

func NewClient(numMsg int, msgSize int, start time.Time) *Client {
	return &Client{
		Sample: Sample{
			JobMsgCnt: uint64(numMsg),
			Start:     start,
		},
		numMsg:  numMsg,
		msgSize: msgSize,
	}
}

func (c *Client) runPublisher(cc *grpc.ClientConn, sample *Sample) {
	stream, err := pb.NewGreeterClient(cc).SayHello(context.Background())
	if err != nil {
		panic(err)
	}
	defer stream.CloseSend()

	req := &pb.HelloRequest{Message: make([]byte, *size)}
	for i := 0; i < c.numMsg; i++ {
		if err := stream.Send(req); err != nil {
			panic(err)
		}

		c.MsgCnt++
		c.MsgBytes += uint64(*size)
	}
	atomic.AddUint64(&sample.MsgCnt, c.MsgCnt)
	atomic.AddUint64(&sample.MsgBytes, c.MsgBytes)
	c.End = time.Now()
}

// IBytes(82854982) -> 79 MiB
func IBytes(s float64) string {
	sizes := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	if s < 10 {
		return fmt.Sprintf("%d B", s)
	}
	e := math.Floor(math.Log(float64(s)) / math.Log(1024))
	suffix := sizes[int(e)]
	val := math.Floor(float64(s)/math.Pow(1024, e)*10+0.5) / 10
	f := "%.0f %s"
	if val < 10 {
		f = "%.1f %s"
	}

	return fmt.Sprintf(f, val, suffix)
}

func msgsPreClient(numMsgs int, numClients int) []int {
	if numMsgs == 0 || numClients == 0 {
		return nil
	}
	results := make([]int, numClients)
	mc := numMsgs / numClients
	for i := 0; i < numClients; i++ {
		results[i] = mc
	}
	extra := numMsgs % numClients
	for i := 0; i < extra; i++ {
		results[i]++
	}
	return results
}

func main() {
	flag.Parse()

	cleanup, err := startServer(*addr)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	log.Printf("Starting gRPC-Go stream benchmark [msgs=%d, msgsize=%s, pubs=%d]\n", *msgs, IBytes(float64(*size)), *pub)

	sample := &Sample{JobMsgCnt: uint64(*msgs)}
	wg := sync.WaitGroup{}
	pubCounts := msgsPreClient(*msgs, *pub)
	clients := make([]*Client, 0, *pub)
	for i := 0; i < *pub; i++ {
		conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		c := NewClient(pubCounts[i], *size, time.Now())
		clients = append(clients, c)

		wg.Add(1)
		go func() {
			defer wg.Done()

			c.runPublisher(conn, sample)
		}()
	}
	wg.Wait()
	for _, client := range clients {
		if sample.Start.IsZero() || sample.Start.After(client.Start) {
			sample.Start = client.Start
		}
		if sample.End.IsZero() || sample.End.Before(client.End) {
			sample.End = client.End
		}
	}

	log.Printf("Stream stats: %d msgs/sec ~ %s/sec (%d msgs)\n", sample.Rate(), IBytes(sample.Throughput()), sample.MsgCnt)
	if len(clients) > 1 {
		for i, c := range clients {
			log.Printf("  [%d] %d msgs/sec ~ %s/sec (%d msgs)\n", i, c.Rate(), IBytes(c.Throughput()), c.MsgCnt)
		}
	}
}
