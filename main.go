package main

import (
	"flag"
	"fmt"
	"sync"

	"github.com/connor-ashton-dev/chat/client"
	"github.com/connor-ashton-dev/chat/server"
)

type Process interface {
	Start()
}

func main() {
	port := flag.Int("port", 1337, "Port number for the client/server to connect to")
	name := flag.String("name", "", "Name to send on client")
	isServer := flag.Bool("server", false, "Run as a server if true, else client")
	flag.Parse()

	var p Process
	if *isServer {
		fmt.Println("Server starting on port:", *port)
		p = server.New(*port)
	} else {
		p = client.New(*port, *name)
	}

	run(p)

}

func run(p Process) {
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		p.Start()
	}()
	wg.Wait()
}
