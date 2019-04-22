package main

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type Server struct {
	name    string
	address *net.TCPAddr
	stop    bool
	group   *sync.WaitGroup
}

type ServerHandler interface {
	Handle(conn net.Conn)
}

func NewServer(name string, address *net.TCPAddr, group *sync.WaitGroup) *Server {
	return &Server{
		name,
		address,
		false,
		group,
	}
}

func (server *Server) listen(handler ServerHandler) {
	server.group.Add(1)
	defer server.group.Done()

	socket, err := net.ListenTCP("tcp", server.address)
	if err != nil {
		log.Fatal(err)
	}
	defer CloseOrFatal(socket)
	log.Printf(`start %s server on %s`, server.name, server.address)

	for {
		if server.stop {
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
		go server.handleConnection(conn, handler)
	}
	log.Printf(`shutdown %s server`, server.name)
}

func (server *Server) handleConnection(conn net.Conn, handler ServerHandler) {
	defer CloseOrLog(conn)
	server.group.Add(1)
	defer server.group.Done()
	log.Printf(`start new connection on %s from %s`, server.name, conn.RemoteAddr())
	handler.Handle(conn)
	log.Printf(`close connection on %s from %s`, server.name, conn.RemoteAddr())
}

func CloseOrFatal(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func CloseOrLog(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		log.Fatal(err)
	}
}
