package mtp

import (
	"github.com/malumar/zoha/api"
	"strings"
)

const (
	OkResponse    = "250 OK"
	OkResponseFmt = "250 OK %s"
)

type CommandHandler func(c *Client, supervisor api.Supervisor, args string)

// RegisteredCommands preserves the order - which makes it possible to find an error
var RegisteredCommands []string

func RegisterCommand(name string, cmd CommandHandler) {
	cname := strings.ToUpper(name)
	RegisteredCommands = append(RegisteredCommands, cname)
	Commands[cname] = cmd
}

var Commands = make(map[string]CommandHandler)
