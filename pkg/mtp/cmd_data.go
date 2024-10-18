package mtp

import (
	"github.com/malumar/zoha/api"
	"time"
)

func init() {
	RegisterCommand("data", cmdData)
}

func cmdData(c *Client, supervisor api.Supervisor, args string) {
	if c.DidHelo() && c.DidMailFrom() && c.DidRecipient() /*&& client.allowNextMessage() */ {
		if err := c.WriteCmd("354 Enter message, ending with . on a line by itself"); err == nil {
			c.state.current = StateData
			c.Con.SetDeadline(time.Now().Add(c.DataStartReceivingTimeoutDuration))
		} else {

		}
	}
}
