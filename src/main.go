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
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	tasks := &sync.WaitGroup{}

	xdebugArgPtr := flag.String("xdebug", "0.0.0.0:9000", "ip:port for xdebug connections")
	registryArgPtr := flag.String("registry", "0.0.0.0:9001", "ip:port for registry connections")
	flag.Parse()

	storage := NewClientList()

	// start registry
	registryAddress, err := net.ResolveTCPAddr("tcp", *registryArgPtr)
	if err != nil {
		log.Fatal(err)
	}
	registryServer := NewServer("registry", registryAddress, tasks)
	registryHandler := &RegistryHandler{storage}
	go registryServer.listen(registryHandler)

	// start pipe
	pipeAddress, err := net.ResolveTCPAddr("tcp", *xdebugArgPtr)
	if err != nil {
		log.Fatal(err)
	}
	proxyServer := NewServer("proxy", pipeAddress, tasks)
	proxyHandler := &ProxyHandler{storage}
	go proxyServer.listen(proxyHandler)

	log.Println("dbgp proxy started")
	// wait os signals
	<-signals
	log.Println("signal Ctrl+C")
	proxyServer.stop = true
	registryServer.stop = true
	tasks.Wait()
	log.Println("dbgp proxy stopped")
}
