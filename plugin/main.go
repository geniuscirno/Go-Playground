package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/geniuscirno/Go-Playground/plugin/plugins"
	"os"
	"plugin"
	"strings"
)

func init() {
	fmt.Println("main init")
}

func loadGreeter(path string) plugins.Greeter {
	p, err := plugin.Open(path)
	if err != nil {
		panic(err)
	}

	v, err := p.Lookup("NewGreeter")
	if err != nil {
		panic(err)
	}

	greeterBuilder := v.(func() plugins.Greeter)
	return greeterBuilder()
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		cmds := strings.Fields(line)
		if len(cmds) == 0 {
			continue
		}
		switch cmds[0] {
		case "load":
			greeter := loadGreeter(cmds[1])
			reply, err := greeter.SayHello(context.Background(), &plugins.HelloRequest{Name: "World"})
			if err != nil {
				panic(err)
			}
			fmt.Println("Reply: ", reply.Message)
		}
	}
}
