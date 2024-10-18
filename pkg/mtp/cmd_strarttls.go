package mtp

import "github.com/malumar/zoha/api"

func init() {
	RegisterCommand("starttls", cmdStartTls)
}

func cmdStartTls(c *Client, supervisor api.Supervisor, args string) {
	if c.Flags.HasFlag(StartTls) {
		c.WriteCmd("220 Go ahead")
	} else {
		if c.DidStartTlsNotAllowed {
			c.ErrorsCount++
		} else {
			c.DidStartTlsNotAllowed = true
			c.WriteCmd("454 TLS not supported")
		}
	}
}
