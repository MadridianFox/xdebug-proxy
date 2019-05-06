package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const InstallCommand = "install"
const ProxyCommand = "proxy"

type ProxyArgs struct {
	XdebugAddr   string
	RegistryAddr string
}

type InstallArgs struct {
	User       string
	Group      string
	OutputFile string
}

type ProxyHandler func(args *ProxyArgs)
type InstallHandler func(args *InstallArgs)

func RunCommand(proxyHandler ProxyHandler, installHandler InstallHandler) {
	proxyCommand := flag.NewFlagSet(ProxyCommand, flag.ExitOnError)
	xdebugAddrPtr := proxyCommand.String("xdebug", "0.0.0.0:9000", "ip:port for xdebug connections")
	registryAddrPtr := proxyCommand.String("registry", "0.0.0.0:9001", "ip:port for idekey registry connections")

	installCommand := flag.NewFlagSet(InstallCommand, flag.ExitOnError)
	userPtr := installCommand.String("user", "www-data", "user for proxy process")
	groupPtr := installCommand.String("group", "www-data", "group for proxy process")
	outputPtr := installCommand.String("out", "/etc/systemd/system/dbgp.service", "path to service unit file")

	if len(os.Args) == 1 {
		fmt.Println("usage: dbgp <command> [<args>]")
		fmt.Println("Available commands: ")
		fmt.Println(" proxy    start dbgp proxy")
		fmt.Println(" install  set dbgp proxy as systemd service")
		return
	}

	switch os.Args[1] {
	case InstallCommand:
		err := installCommand.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
		installHandler(&InstallArgs{
			User:       *userPtr,
			Group:      *groupPtr,
			OutputFile: *outputPtr,
		})
	case ProxyCommand:
		err := proxyCommand.Parse(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
		proxyHandler(&ProxyArgs{
			XdebugAddr:   *xdebugAddrPtr,
			RegistryAddr: *registryAddrPtr,
		})
	}
}
