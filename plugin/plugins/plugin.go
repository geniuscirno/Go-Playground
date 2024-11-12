package plugins

import (
	"context"
	"fmt"
)

func init() {
	fmt.Println("plugins init1")
}

type HelloRequest struct {
	Name string
}

type HelloReply struct {
	Message string
}

type Greeter interface {
	SayHello(ctx context.Context, in *HelloRequest) (*HelloReply, error)
}
