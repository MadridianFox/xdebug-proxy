package main

import (
	"dbgp-proxy/src/proxy"
	"log"
)

func main() {
	RunCommand(HandleProxyCommand, HandleInstallCommand)
}

func HandleProxyCommand(proxyArgs *ProxyArgs) {
	proxy.Run(proxyArgs.RegistryAddr, proxyArgs.XdebugAddr)
}

func HandleInstallCommand(installArgs *InstallArgs) {
	err := SaveConfig(installArgs)
	if err != nil {
		log.Fatal(err)
	}
}
