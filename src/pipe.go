package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"time"
)

type ProxyHandler struct {
	clients *ClientList
}

func (proxy *ProxyHandler) Handle(conn net.Conn) {
	reader := bufio.NewReader(conn)
	length, err := reader.ReadBytes(0)
	if err != nil {
		if err != io.EOF {
			log.Print(err)
		}
	}
	message, err := reader.ReadBytes(0)
	if err != nil {
		if err != io.EOF {
			log.Print(err)
		}
	}
	idekey, err := GetIdekey(message[:len(message)-1])
	if err != nil {
		log.Println(err)
	}
	log.Printf("idekey: %s\n", idekey)
	err = proxy.sendAndPipe(conn, idekey, append(length, message...))
	if err != nil {
		log.Println(err)
	}
}

func (proxy *ProxyHandler) sendAndPipe(server net.Conn, idekey string, initMessage []byte) error {
	clientAddress, ok := proxy.clients.FindClient(idekey)
	if !ok {
		return errors.New(`client with idekey "` + idekey + `" isn't registered`)
	}
	log.Printf("send init packet to: %s\n", clientAddress.fullAddress())
	client, err := net.Dial("tcp", clientAddress.fullAddress())
	defer CloseOrLog(client)
	if err != nil {
		return err
	}
	log.Println("IDE Connected")
	_, err = io.Copy(client, bytes.NewReader(initMessage))
	if err != nil {
		return err
	}
	log.Println("init accepted")
	clientChan := make(chan error)
	serverChan := make(chan error)
	go func() {
		_, err = io.Copy(client, server)
		clientChan <- err
	}()

	go func() {
		_, err = io.Copy(server, client)
		serverChan <- err
	}()
LOOP:
	for {
		select {
		case err = <-clientChan:
			log.Println("IDE connection closed")
			break LOOP
		case err = <-serverChan:
			log.Println("XDebug connection closed")
			break LOOP
		default:
			time.Sleep(time.Millisecond * 50)
		}
	}
	log.Println("stop piping")
	return nil
}
