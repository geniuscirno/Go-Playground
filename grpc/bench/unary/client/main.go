package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	pb "github.com/geniuscirno/Go-Playground/grpc/bench/unary/helloworld"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

const (
	B  = 1
	KB = B << 10
	MB = KB << 10
	GB = MB << 10
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	msgs = flag.Int("msgs", 100000, "the message to send")
	size = flag.Int("size", 128, "the size of each message")
	pub  = flag.Int("pub", 1, "number of concurrent publishers")
)

type Client struct {
	c       pb.GreeterClient
	startAt time.Time
	entAt   time.Time
	msgs    int
}

func (c *Client) publish(message string) {
	_, err := c.c.SayHello(context.TODO(), &pb.HelloRequest{Message: message})
	if err != nil {
		panic(err)
	}
}

func (c *Client) Seconds() float64 {
	return c.entAt.Sub(c.startAt).Seconds()
}

func byteSize(size int) string {
	switch {
	case size > GB:
		return fmt.Sprintf("%.2fGB", float64(size/GB)+float64(size%GB)/float64(GB))
	case size > MB:
		return fmt.Sprintf("%.2fMB", float64(size/MB)+float64(size%MB)/float64(MB))
	case size > KB:
		return fmt.Sprintf("%.2fKB", float64(size/KB)+float64(size%KB)/float64(KB))
	default:
		return fmt.Sprintf("%dB", size)
	}
}

func genMessage(size int) string {
	sb := strings.Builder{}
	for i := 0; i < size; i++ {
		sb.WriteByte(letters[rand.Intn(len(letters))])
	}
	return sb.String()
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
	// Set up a connection to the server.
	log.Printf("Starting gRPC-Go unary benchmark [msgs=%d, msgsize=%s, pubs=%d]\n", *msgs, byteSize(*size), *pub)

	message := genMessage(*size)

	var (
		startAt time.Time
		endAt   time.Time
	)
	mu := &sync.Mutex{}
	wg := sync.WaitGroup{}
	pubCounts := msgsPreClient(*msgs, *pub)
	clients := make([]*Client, 0, *pub)
	for i := 0; i < *pub; i++ {
		conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		c := &Client{
			c:       pb.NewGreeterClient(conn),
			startAt: time.Now(),
			msgs:    pubCounts[i],
		}
		clients = append(clients, c)
		mu.Lock()
		if startAt.IsZero() {
			startAt = c.startAt
		}
		mu.Unlock()
		wg.Add(1)
		go func() {
			for j := 0; j < c.msgs; j++ {
				c.publish(message)
			}
			c.entAt = time.Now()
			mu.Lock()
			if endAt.Before(c.entAt) {
				endAt = c.entAt
			}
			mu.Unlock()
			wg.Done()
			conn.Close()
		}()
	}
	wg.Wait()

	duration := endAt.Sub(startAt).Seconds()
	log.Printf("Unary stats: %d msgs/sec ~ %s/sec\n", int(float64(*msgs)/duration), byteSize(int(float64(*size**msgs)/duration)))
	if len(clients) > 1 {
		for i, c := range clients {
			log.Printf("  [%d] %d msgs/sec ~ %s/sec (%d msgs)\n", i, int(float64(c.msgs)/c.Seconds()), byteSize(int(float64(*size*c.msgs)/c.Seconds())), c.msgs)
		}
	}
}
