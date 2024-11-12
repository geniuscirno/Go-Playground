package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/geniuscirno/Go-Playground/plugin/plugins"
	"github.com/geniuscirno/Go-Playground/plugin/plugins/plugin1"
	"log"
	"os"
	"path/filepath"
	"plugin"
	"strings"
)

func init() {
	fmt.Println("main init")
}

func loadGreeter(path string) plugins.Greeter {
	p, err := plugin.Open(path)
	if err != nil {
		log.Println(err)
		return nil
	}

	v, err := p.Lookup("NewGreeter")
	if err != nil {
		log.Println(err)
		return nil
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
			greeter := loadGreeter(filepath.Join("./plugins", cmds[1]))
			if greeter == nil {
				continue
			}
			reply, err := greeter.SayHello(context.Background(), &plugins.HelloRequest{Name: "World"})
			if err != nil {
				panic(err)
			}
			fmt.Println("Reply: ", reply.Message)
		case "local":
			greeter := plugin1.NewGreeter()
			reply, err := greeter.SayHello(context.Background(), &plugins.HelloRequest{Name: "World"})
			if err != nil {
				panic(err)
			}
			fmt.Println("Reply: ", reply.Message)
		}
	}
}
