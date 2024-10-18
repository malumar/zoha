package mtp

import "github.com/malumar/zoha/api"

func init() {
	RegisterCommand("quit", cmdQuit)
}

func cmdQuit(c *Client, supervisor api.Supervisor, args string) {
	c.WeClose = true
	c.state.disconnectReason = "quit command"
	c.properClose = true
	c.WriteCmd("221 Bye")
}
