package main

import (
	"context"
	"fmt"
	"github.com/geniuscirno/Go-Playground/plugin/plugins"
)

func init() {
	fmt.Println("plugin2 init")
}

//go:generate go build -buildmode=plugin -o plugin2.so plugin2.go
type helloWorld struct{}

func NewGreeter() plugins.Greeter {
	return &helloWorld{}
}

func (h *helloWorld) SayHello(ctx context.Context, in *plugins.HelloRequest) (*plugins.HelloReply, error) {
	return &plugins.HelloReply{Message: "plugin2: Hello, " + in.Name}, nil
}
