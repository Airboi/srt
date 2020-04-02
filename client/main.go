package main

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
	"os"
	"srt/client/go-socks5"
	"time"
)

var chanExit = make(chan int)

const (
	CHAN_FORWARD   = "RbgEySPMPi"
	CHAN_HEARTBEAT = "uSYeIbUQoR"
	CHAN_COMMAND   = "rIHqXLCqRN"
	COMMAND_KILL   = "aAjcDqEIvI"
)

type Config struct {
	SSHHost   string
	SSHPort   string
	SSHUser   string
	SSHPwd    string
	SocksUser string
	SocksPwd  string
}

func (c *Config) ServerString() string {
	return fmt.Sprintf("%s:%s", c.SSHHost, c.SSHPort)
}

type Client struct {
	Config     *Config
	SockServer *socks5.Server
	ServerAddr net.Addr
}

func (c *Client) connect() {
	var client net.Conn
	var err error
	for {
		client, err = net.Dial("tcp", c.Config.ServerString())
		if err == nil {
			break
		}
		time.Sleep((3 * time.Second))
	}

	config := &ssh.ClientConfig{
		User:            c.Config.SSHUser,
		Auth:            []ssh.AuthMethod{ssh.Password(c.Config.SSHPwd)},
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
	case CHAN_FORWARD:
		c.SockServer.ServeConn(channel, c.ServerAddr)
	case CHAN_HEARTBEAT:
		go func() {
			bufConn := bufio.NewReader(channel)
			tcommand := make([]byte, 4)
			for {
				if _, err := bufConn.Read(tcommand); err != nil {
					c.connect()
					return
				}
			}
		}()
	case CHAN_COMMAND:
		go func() {
			bufConn := bufio.NewReader(channel)
			tcommand := make([]byte, 256)
			for {
				n, err := bufConn.Read(tcommand)
				if err == nil {
					if string(tcommand[:n]) == COMMAND_KILL {
						chanExit <- 1
						return
					}
				} else {
					return
				}
			}
		}()
	}
}

func (c *Client) startSock() {
	cred := socks5.StaticCredentials{
		c.Config.SocksUser: c.Config.SocksPwd,
	}
	cator := socks5.UserPassAuthenticator{Credentials: cred}
	conf := &socks5.Config{AuthMethods: []socks5.Authenticator{cator}}
	server, err := socks5.New(conf)
	if err != nil {
		chanExit <- 1
		return
	}
	c.SockServer = server
}

func main() {
	if len(os.Args) != 7 {
		return
	}
	c := &Client{
		Config: &Config{
			SSHHost:   os.Args[1],
			SSHPort:   os.Args[2],
			SSHUser:   os.Args[3],
			SSHPwd:    os.Args[4],
			SocksUser: os.Args[5],
			SocksPwd:  os.Args[6],
		},
	}
	c.startSock()
	c.connect()
	<-chanExit
}
