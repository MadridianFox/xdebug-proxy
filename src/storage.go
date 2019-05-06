package main

import (
	"net"
	"sync"
)

type DbgpClient struct {
	address string
	port    string
	idekey  string
}

func (client *DbgpClient) fullAddress() string {
	return net.JoinHostPort(client.address, client.port)
}

type ClientList struct {
	sync.Mutex
	clients map[string]*DbgpClient
}

func NewClientList() *ClientList {
	return &ClientList{clients: map[string]*DbgpClient{}}
}

func (list *ClientList) setFromConfig(items []ConfigItem) error {
	for _, item := range items {
		host, port, err := net.SplitHostPort(item.Address)
		if err != nil {
			return err
		}
		list.AddClient(&DbgpClient{host, port, item.Idekey})
	}
	return nil
}

func (list *ClientList) AddClient(client *DbgpClient) {
	list.Lock()
	defer list.Unlock()
	list.clients[client.idekey] = client
}

func (list *ClientList) DeleteClient(idekey string) {
	list.Lock()
	defer list.Unlock()
	delete(list.clients, idekey)
}

func (list *ClientList) FindClient(idekey string) (*DbgpClient, bool) {
	list.Lock()
	defer list.Unlock()
	client, ok := list.clients[idekey]
	return client, ok
}

func (list *ClientList) HasClientWithPort(port string) bool {
	var portInUse bool
	for _, client := range list.clients {
		if client.port == port {
			portInUse = true
			break
		}
	}
	return portInUse
}
