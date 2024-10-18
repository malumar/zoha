package mtp

import "github.com/malumar/zoha/api"

func init() {
	RegisterCommand("nop", cmdNop)
}

func cmdNop(c *Client, supervisor api.Supervisor, args string) {
	c.WriteCmd("250 OK")
}
