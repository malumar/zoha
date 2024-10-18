package mtp

import (
	"fmt"
	"github.com/malumar/zoha/api"
	"strings"
)

func init() {
	RegisterCommand("helo", cmdHelo)
}

func cmdHelo(c *Client, supervisor api.Supervisor, args string) {
	//if !client.DidHeloCmd {
	c.Helo = strings.Trim(" ", args)
	c.WriteCmd(fmt.Sprintf("250 %s", c.ListenerReverseHostname))
	c.DidHeloCmd = true

}
