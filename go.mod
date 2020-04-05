module srt

go 1.13

require (
	github.com/mitchellh/gox v1.0.1 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	golang.org/x/crypto v0.0.0-20200403201458-baeed622b8d8
	srt/client/go-socks5 v0.0.0-00010101000000-000000000000
)

replace srt/client/go-socks5 => ./client/go-socks5
