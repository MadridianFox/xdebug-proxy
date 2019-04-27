package main

import (
	"dbgp-proxy/src/proxy"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
)

func main() {
	RunCommand(HandleProxyCommand, HandleInstallCommand)
}

func HandleProxyCommand(args *ProxyArgs) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	tasks := &sync.WaitGroup{}

	storage := proxy.NewClientList()

	// start registry
	registryAddress, err := net.ResolveTCPAddr("tcp", args.RegistryAddr)
	if err != nil {
		log.Fatal(err)
	}
	registryServer := proxy.NewServer("registry", registryAddress, tasks)
	registryHandler := proxy.NewRegistryHandler(storage)
	go registryServer.Listen(registryHandler)

	// start pipe
	pipeAddress, err := net.ResolveTCPAddr("tcp", args.XdebugAddr)
	if err != nil {
		log.Fatal(err)
	}
	proxyServer := proxy.NewServer("proxy", pipeAddress, tasks)
	proxyHandler := proxy.NewPipeHandler(storage)
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

//noinspection ALL
func HandleInstallCommand(agrs *InstallArgs) {

}
