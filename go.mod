module srt

go 1.13

require (
	github.com/chzyer/logex v1.1.10 // indirect
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/chzyer/test v0.0.0-20180213035817-a1ea475d72b1 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	golang.org/x/crypto v0.0.0-20200403201458-baeed622b8d8
	srt/client/go-socks5 v0.0.0-00010101000000-000000000000
)

replace srt/client/go-socks5 => ./client/go-socks5
