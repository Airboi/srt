module srt

go 1.13

require (
	github.com/chzyer/readline v0.0.0-20180603132655-2972be24d48e
	github.com/jroimartin/gocui v0.4.0
	github.com/manifoldco/promptui v0.7.0
	github.com/marcusolsson/tui-go v0.4.0
	github.com/mitchellh/gox v1.0.1 // indirect
	github.com/nsf/termbox-go v0.0.0-20200204031403-4d2b513ad8be
	github.com/olekukonko/tablewriter v0.0.4
	github.com/urfave/cli/v2 v2.2.0
	github.com/wagoodman/keybinding v0.0.0-20181213133715-6a824da6df05
	golang.org/x/crypto v0.0.0-20200403201458-baeed622b8d8
	srt/client/go-socks5 v0.0.0-00010101000000-000000000000
)

replace srt/client/go-socks5 => ./client/go-socks5
