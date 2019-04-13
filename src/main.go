package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
)

func main() {
	log.Println("dbgp proxy started")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	tasks := &sync.WaitGroup{}

	xdebugArgPtr := flag.String("xdebug", "0.0.0.0:9000", "ip:port for xdebug connections")
	registryArgPtr := flag.String("registry", "0.0.0.0:9001", "ip:port for registry connections")
	flag.Parse()

	// start registry
	registryAddress, err := net.ResolveTCPAddr("tcp", *registryArgPtr)
	if err != nil {
		log.Fatal(err)
	}
	registry := NewProxyRegistry(registryAddress)
	go registry.listen(tasks)

	// start pipe
	pipeAddress, err := net.ResolveTCPAddr("tcp", *xdebugArgPtr)
	if err != nil {
		log.Fatal(err)
	}
	pipe := NewPipe(pipeAddress, &registry.clients)
	go pipe.listen(tasks)

	// wait os signals
	<-signals
	log.Println("signal Ctrl+C")
	pipe.stopping = true
	registry.stopping = true
	tasks.Wait()
	log.Println("dbgp proxy stopped")
}
