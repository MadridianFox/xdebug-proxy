package main

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"sync"
	"time"
)

type ProxyXmlMessage struct {
	XMLName xml.Name
	Success int8   `xml:"success,attr"`
	Idekey  string `xml:"idekey,attr"`
	Error   string `xml:"error>message"`
}

type RegistryClient struct {
	address string
	port    string
	idekey  string
}

func (client *RegistryClient) fullAddress() string {
	return net.JoinHostPort(client.address, client.port)
}

type RegistryClients map[string]RegistryClient

type ProxyRegistry struct {
	address  *net.TCPAddr
	clients  RegistryClients
	stopping bool
}

func NewProxyRegistry(address *net.TCPAddr) *ProxyRegistry {
	return &ProxyRegistry{
		address,
		RegistryClients{},
		false,
	}
}

func (registry *ProxyRegistry) hasClientWithPort(port string) bool {
	var portInUse bool
	for _, client := range registry.clients {
		if client.port == port {
			portInUse = true
			break
		}
	}
	return portInUse
}

func (registry *ProxyRegistry) listen(tasks *sync.WaitGroup) {
	tasks.Add(1)
	defer tasks.Done()
	socket, err := net.ListenTCP("tcp", registry.address)
	if err != nil {
		log.Fatal(err)
	}
	defer CloseOrFatal(socket)
	log.Printf(`listen registry connections on %s`, registry.address.String())
	for {
		if registry.stopping {
			break
		}
		_ = socket.SetDeadline(time.Now().Add(time.Second * 2))
		conn, err := socket.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			log.Print(err)
			continue
		}
		go registry.handleConnection(conn, tasks)
	}
	log.Println("shutdown registry listener")
}

func (registry *ProxyRegistry) handleConnection(conn net.Conn, tasks *sync.WaitGroup) {
	defer CloseOrLog(conn)
	tasks.Add(1)
	defer tasks.Done()
	r := bufio.NewReader(conn)
	log.Println("new registry connection")
	data, err := r.ReadBytes(0)
	if err != nil {
		if err != io.EOF {
			log.Print(err)
		}
	}
	message := string(data[:len(data)-1])
	err = registry.processMessage(message, conn)
	if err != nil {
		log.Print(err)
	}
}

func extractByRegexp(pattern string, data string) (string, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	groups := re.FindStringSubmatch(data)
	if len(groups) < 2 {
		return "", errors.New(fmt.Sprintf(`no groups matched in "%s" by pattern "%s"`, data, pattern))
	}
	return groups[1], nil
}

func (registry *ProxyRegistry) sendMessage(conn net.Conn, messageType string, idekey string, error string) error {
	var success int8
	if error == "" {
		success = 1
	} else {
		success = 0
	}
	message := ProxyXmlMessage{
		XMLName: xml.Name{Local: messageType},
		Success: success,
		Idekey:  idekey,
		Error:   error,
	}
	xmlMessage, err := xml.Marshal(message)
	if err != nil {
		return err
	}
	_, err = io.Copy(conn, bytes.NewReader(xmlMessage))
	if err != nil {
		return err
	}
	return nil
}

func (registry *ProxyRegistry) processMessage(message string, conn net.Conn) error {
	command, err := extractByRegexp(`^([^\s]+)`, message)
	if err != nil {
		_ = registry.sendMessage(conn, command, "", "Can't parse command.")
		return err
	}
	idekey, err := extractByRegexp(`-k ([^\s]+)`, message)
	if err != nil {
		_ = registry.sendMessage(conn, command, idekey, "Can't parse idekey.")
		return err
	}
	if command == "proxyinit" {
		port, err := extractByRegexp(`-p ([^\s]+)`, message)
		if err != nil {
			_ = registry.sendMessage(conn, command, idekey, "Can't parse port.")
			return err
		}
		if registry.hasClientWithPort(port) {
			_ = registry.sendMessage(conn, command, idekey, fmt.Sprintf(`Port "%s" already in use.`, port))
			return nil
		}
		host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err != nil {
			_ = registry.sendMessage(conn, command, idekey, "Internal error.")
			return err
		}
		registry.clients[idekey] = RegistryClient{
			idekey:  idekey,
			address: host,
			port:    port,
		}
		log.Println(fmt.Sprintf(`add client with host "%s", port "%s" and idekey "%s"`, host, port, idekey))
	} else if command == "proxystop" {
		_, ok := registry.clients[idekey]
		if ok {
			delete(registry.clients, idekey)
			log.Println(fmt.Sprintf(`delete client with idekey "%s"`, idekey))
		} else {
			log.Println(fmt.Sprintf(`attempt to delete idekey "%s"`, idekey))
			err = registry.sendMessage(conn, command, idekey, fmt.Sprintf(`Idekey "%s" isn't registered'`, idekey))
			if err != nil {
				return err
			}
			return nil
		}
	} else {
		return errors.New(fmt.Sprintf(`undefined command "%s"`, command))
	}
	err = registry.sendMessage(conn, command, idekey, "")
	if err != nil {
		return err
	}
	return nil
}
