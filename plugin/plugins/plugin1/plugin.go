package plugin1

import (
	"context"
	"fmt"
	"github.com/geniuscirno/Go-Playground/plugin/plugins"
	"log"
)

var X int32

const Version = 2

func init() {
	log.Printf("plugin1 init, version=%d", Version)
}

//go:generate go build -buildmode=plugin -o plugin1.so plugin1.go
type helloWorld struct{}

func NewGreeter() plugins.Greeter {
	return &helloWorld{}
}

func (h *helloWorld) SayHello(ctx context.Context, in *plugins.HelloRequest) (*plugins.HelloReply, error) {
	X++
	if X > 5 {
		panic("xx")
	}
	return &plugins.HelloReply{Message: fmt.Sprintf("plugin1: Hello, %s @%d, version=%d", in.Name, X, Version)}, nil
}
