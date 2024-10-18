package mtp

import (
	"fmt"
	"github.com/malumar/zoha/api"
)

func init() {
	// if smtp
	RegisterCommand("ehlo", cmdEhlo)
	// if lmtp
	RegisterCommand("lhlo", cmdEhlo)
}

func cmdEhlo(c *Client, supervisor api.Supervisor, args string) {
	//if !client.DidHeloCmd~ {
	c.UseEhlo = true
	c.Helo = args
	c.DidHeloCmd = true
	response := fmt.Sprintf("250-%s at your service, [%s]", c.ListenerReverseHostname, c.Info.RemoteAddr)

	if c.MaxMessageSizeInBytes > 0 {
		response += fmt.Sprintf("\n250-SIZE %d", c.MaxMessageSizeInBytes)
	}

	if c.Flags.HasFlag(StartTls) {
		response += "\n250STARTTLS"
	}

	if c.Flags.HasFlag(Authorization) {
		response += "\n250-AUTH PLAIN LOGIN\n250-SMTPUTF8\n250-8BITMIME\n250 PIPELING"
	} else {
		response += "\n250-SMTPUTF8\n250-8BITMIME\n250 PIPELING"

	}
	c.WriteCmd(response)

}
