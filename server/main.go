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
	"time"
)

type Endpoint struct {
	Host string
	Port string
	User string
	Pwd  string
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%s", endpoint.Host, endpoint.Port)
}

type Session struct {
	SessionId    int
	sshConn      *ssh.ServerConn
	Disconnected bool
}

type Server struct {
	SSHServer        *Endpoint
	ProxyServer      *Endpoint
	Sessions         []Session
	CurrentSessionId int
	LastSessionId    int
}

func (s *Server) startProxy() {
	listener, err := net.Listen("tcp", s.ProxyServer.String())
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
			if c.User() == s.SSHServer.User && string(pass) == s.SSHServer.Pwd {
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

	listener, err := net.Listen("tcp", s.SSHServer.String())
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
		s.LastSessionId += 1
		if s.CurrentSessionId == -1 {
			s.CurrentSessionId = sess.SessionId
		}
		s.Sessions = append(s.Sessions, sess)
		log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		go ssh.DiscardRequests(reqs)
		go s.keepAlive(sess.SessionId)
		go handleChannels(chans)
	}
}

func (s *Server) keepAlive(id int) {
	newChan, _, err := s.Sessions[id].sshConn.OpenChannel("heartBeat", []byte(s.SSHServer.String()))
	if err != nil {
		s.Sessions[id].Disconnected = true
		return
	}
	for {
		time.Sleep(10 * time.Second)
		_, err := newChan.Write([]byte("ping"))
		if err != nil {
			s.Sessions[id].Disconnected = true
			s.CurrentSessionId = -1
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
	if s.CurrentSessionId == -1 {
		return
	}
	newChan, _, err := s.Sessions[s.CurrentSessionId].sshConn.OpenChannel("forward", []byte(s.SSHServer.String()))
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
	if len(os.Args) != 6 {
		fmt.Println("Usage: ./server sshAddr sshPort  proxyAddr proxyPort password")
		fmt.Println("Ex: ./server 0.0.0.0 2222 0.0.0.0 8080 thisrandomkey")
		return
	}

	s := Server{
		SSHServer: &Endpoint{
			Host: os.Args[1],
			Port: os.Args[2],
			User: "tom",
			Pwd:  os.Args[5],
		},
		ProxyServer: &Endpoint{
			Host: os.Args[3],
			Port: os.Args[4],
			User: "",
			Pwd:  "",
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
				data := []string{strconv.Itoa(session.SessionId), session.sshConn.RemoteAddr().String()}
				datas = append(datas, data)
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"Id", "Address"})
			for _, v := range datas {
				table.Append(v)
			}
			table.Render()
			fmt.Println("current session: " + strconv.Itoa(s.CurrentSessionId))
			continue
		}
		if cmd[0] == "use" {
			if len(cmd) == 2 {
				if id, err := strconv.Atoi(cmd[1]); err == nil {
					if id <= s.LastSessionId && id > -1 {
						if s.Sessions[id].Disconnected {
							fmt.Println("that session was disconnected")
							continue
						}
						s.CurrentSessionId = id
					}
				}
			}
			continue
		}
	}
}
