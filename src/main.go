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
	//registry := NewProxyRegistry(registryAddress)
	registryServer := NewServer("registry", registryAddress, tasks)
	registryHandler := &RegistryHandler{RegistryClients{}}
	go registryServer.listen(registryHandler)

	// start pipe
	pipeAddress, err := net.ResolveTCPAddr("tcp", *xdebugArgPtr)
	if err != nil {
		log.Fatal(err)
	}
	proxyServer := NewServer("proxy", pipeAddress, tasks)
	proxyHandler := &ProxyHandler{&registryHandler.clients}
	go proxyServer.listen(proxyHandler)

	// wait os signals
	<-signals
	log.Println("signal Ctrl+C")
	proxyServer.stop = true
	registryServer.stop = true
	tasks.Wait()
	log.Println("dbgp proxy stopped")
}
