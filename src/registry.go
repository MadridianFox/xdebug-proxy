package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

type RegistryHandler struct {
	clients *ClientList
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

func (registry *RegistryHandler) sendMessage(conn net.Conn, messageType string, idekey string, error string) error {
	message, err := CreateInitMessage(messageType, idekey, error)
	if err != nil {
		return err
	}
	_, err = io.Copy(conn, bytes.NewReader(message))
	if err != nil {
		return err
	}
	return nil
}

func (registry *RegistryHandler) prepareErrorMessage(conn net.Conn, command *IdeCommand) func(message string) {
	return func(messge string) {
		_ = registry.sendMessage(conn, command.Name, command.Idekey, messge)
	}
}

//noinspection GoNilness
func (registry *RegistryHandler) processMessage(message string, conn net.Conn) error {
	command, err := ParseIdeCommand(message)
	report := registry.prepareErrorMessage(conn, command)
	if err != nil {
		report("invalid command")
	}
	if command.Name == CommandInit {
		if registry.clients.HasClientWithPort(command.Port) {
			report(fmt.Sprintf(`port "%s" already in use.`, command.Port))
			return nil
		}
		host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err != nil {
			report("internal error.")
			return err
		}
		registry.clients.AddClient(&DbgpClient{idekey: command.Idekey, address: host, port: command.Port})
		log.Println(fmt.Sprintf(`add client with host "%s", port "%s" and idekey "%s"`, host, command.Port, command.Idekey))
	} else if command.Name == CommandStop {
		_, ok := registry.clients.FindClient(command.Idekey)
		if ok {
			registry.clients.DeleteClient(command.Idekey)
			log.Println(fmt.Sprintf(`delete client with idekey "%s"`, command.Idekey))
		} else {
			log.Println(fmt.Sprintf(`attempt to delete idekey "%s"`, command.Idekey))
			report(fmt.Sprintf(`idekey "%s" isn't registered'`, command.Idekey))
			return nil
		}
	} else {
		return errors.New(fmt.Sprintf(`undefined command "%s"`, command))
	}
	err = registry.sendMessage(conn, command.Name, command.Idekey, "")
	if err != nil {
		return err
	}
	return nil
}
