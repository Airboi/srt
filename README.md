srt
=========

SSH Reverse Tunnel

![alt text](https://i.imgur.com/jwJ9sC3.png)

---------
****Build****
 - edit config in Makefile:
 ```
 SSHHost = your-ssh-server-ip
 SSHPort = your-ssh-server-port
 SSHUser = your-ssh-server-user
 SSHPwd = your-ssh-server-password
 SocksUser = your-socks-server-user
 SocksPwd = your-socks-server-pwd
 Tag = client-tag
 ```
 - ~#:make dep
 - ~#:make build

---------
****Usage****

 - Server: ./server.elf 0.0.0.0 2222 userssh passssh 0.0.0.0 8080
 
 - Target: client.exe
 
 - Attacker: set proxy socks5 ServerIp:8080 with usersocks/passsocks
 
 ---------
 
****Ref****
 
 - http://blog.ralch.com/tutorial/golang-ssh-tunneling/
 - https://blog.gopheracademy.com/go-and-ssh/
 - https://github.com/armon/go-socks5