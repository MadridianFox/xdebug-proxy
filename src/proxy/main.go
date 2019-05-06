package proxy

import (
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func Run(registryAddr string, xdebugAddr string) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	tasks := &sync.WaitGroup{}

	storage := NewClientList()

	// start registry
	registryAddress, err := net.ResolveTCPAddr("tcp", registryAddr)
	if err != nil {
		log.Fatal(err)
	}
	registryServer := NewServer("registry", registryAddress, tasks)
	registryHandler := NewRegistryHandler(storage)
	go registryServer.Listen(registryHandler)

	// start pipe
	pipeAddress, err := net.ResolveTCPAddr("tcp", xdebugAddr)
	if err != nil {
		log.Fatal(err)
	}
	proxyServer := NewServer("proxy", pipeAddress, tasks)
	proxyHandler := NewPipeHandler(storage)
	go proxyServer.Listen(proxyHandler)

	log.Println("dbgp proxy started")
	// wait os signals
	<-signals
	log.Println("signal Ctrl+C")
	proxyServer.Stop()
	registryServer.Stop()
	tasks.Wait()
	log.Println("dbgp proxy stopped")
}
