package main

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
	"os"
	"client/go-socks5"
	"time"
)

var chanExit = make(chan int)

type Endpoint struct {
	Host string
	Port string
	User string
	Pwd  string
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%s", endpoint.Host, endpoint.Port)
}

type Client struct {
	SSHClient  *Endpoint
	SockServer *socks5.Server
	ServerAddr net.Addr
}

func (c *Client) connect() {
	var client net.Conn
	var err error
	for {
		client, err = net.Dial("tcp", c.SSHClient.String())
		if err == nil {
			break
		}
		time.Sleep((3 * time.Second))
	}

	config := &ssh.ClientConfig{
		User:            c.SSHClient.User,
		Auth:            []ssh.AuthMethod{ssh.Password(c.SSHClient.Pwd)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshConn, chans, reqs, err := ssh.NewClientConn(client, "", config)
	if err != nil {
		chanExit <- 1
		return
	}
	c.ServerAddr = sshConn.RemoteAddr()
	go ssh.DiscardRequests(reqs)
	c.handleChannels(chans)
}

func (c *Client) handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go c.handleChannel(newChannel)
	}
}

func (c *Client) handleChannel(newChannel ssh.NewChannel) {

	channel, requests, err := newChannel.Accept()
	if err != nil {
		return
	}
	go ssh.DiscardRequests(requests)
	chanType := newChannel.ChannelType()
	switch chanType {
	case "forward":
		c.SockServer.ServeConn(channel, c.ServerAddr)
	case "heartBeat":
		go func() {
			bufConn := bufio.NewReader(channel)
			tcommand := make([]byte, 4)
			for {
				if _, err := bufConn.Read(tcommand); err != nil {
					chanExit <- 1
					return
				}
			}
		}()
	}

}

func (c *Client) startSock() {
	conf := &socks5.Config{}
	server, err := socks5.New(conf)
	if err != nil {
		chanExit <- 1
		return
	}
	c.SockServer = server
}

func main() {
	if len(os.Args) != 4 {
		return
	}
	c := &Client{
		SSHClient: &Endpoint{
			Host: os.Args[1],
			Port: os.Args[2],
			User: "tom",
			Pwd:  os.Args[3],
		},
	}
	c.startSock()
	c.connect()
	<-chanExit
}
