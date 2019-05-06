package main

import (
	"flag"
	"io/ioutil"
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

	configPathPtr := flag.String("config", "/etc/dbgp/config.yml", "path to config file")
	flag.Parse()

	configData, err := ioutil.ReadFile(*configPathPtr)
	if err != nil {
		log.Fatal(err)
	}
	config, err := parseConfig(configData)
	if err != nil {
		log.Fatal(err)
	}

	storage := NewClientList()
	err = storage.setFromConfig(config.Predefined)
	if err != nil {
		log.Fatal(err)
	}
	// start registry
	registryAddress, err := net.ResolveTCPAddr("tcp", config.RegistryAddr)
	if err != nil {
		log.Fatal(err)
	}
	registryServer := NewServer("registry", registryAddress, tasks)
	registryHandler := &RegistryHandler{storage}
	go registryServer.listen(registryHandler)

	// start pipe
	pipeAddress, err := net.ResolveTCPAddr("tcp", config.XdebugAddr)
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
