package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type ProxyPipe struct {
	address  *net.TCPAddr
	clients  *RegistryClients
	stopping bool
}

func NewPipe(address *net.TCPAddr, clients *RegistryClients) *ProxyPipe {
	return &ProxyPipe{
		address,
		clients,
		false,
	}
}

func (proxy *ProxyPipe) listen(tasks *sync.WaitGroup) {
	tasks.Add(1)
	defer tasks.Done()
	socket, err := net.ListenTCP("tcp", proxy.address)
	if err != nil {
		log.Fatal(err)
	}
	defer CloseOrFatal(socket)
	log.Printf(`listen xdebug connections on %s`, proxy.address.String())
	for {
		if proxy.stopping {
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
		go proxy.handleConnection(conn, tasks)
	}

	log.Println("shutdown xdebug listener")
}

func (proxy *ProxyPipe) handleConnection(conn net.Conn, tasks *sync.WaitGroup) {
	defer CloseOrLog(conn)
	tasks.Add(1)
	defer tasks.Done()
	r := bufio.NewReader(conn)
	log.Println("new xdebug connection")

	length, err := r.ReadBytes(0)
	if err != nil {
		if err != io.EOF {
			log.Print(err)
		}
	}
	message, err := r.ReadBytes(0)
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
	log.Println("connection closed")
}

func (proxy *ProxyPipe) sendAndPipe(server net.Conn, idekey string, initMessage []byte) error {
	client_address, ok := (*proxy.clients)[idekey]
	if !ok {
		return errors.New(`client with idekey "` + idekey + `" isn't registered`)
	}
	log.Printf("send init packet to: %s\n", client_address.fullAddress())
	client, err := net.Dial("tcp", client_address.fullAddress())
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
