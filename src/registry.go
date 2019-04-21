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

func (registry *RegistryHandler) hasClientWithPort(port string) bool {
	var portInUse bool
	for _, client := range registry.clients {
		if client.port == port {
			portInUse = true
			break
		}
	}
	return portInUse
}

type RegistryHandler struct {
	clients RegistryClients
}

func (registry *RegistryHandler) Handle(conn net.Conn) {
	reader := bufio.NewReader(conn)
	log.Println("new registry connection")
	data, err := reader.ReadBytes(0)
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

func (registry *RegistryHandler) sendMessage(conn net.Conn, messageType string, idekey string, error string) error {
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

func (registry *RegistryHandler) processMessage(message string, conn net.Conn) error {
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
