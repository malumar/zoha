package mtp

import "github.com/malumar/zoha/api"

func init() {
	RegisterCommand("rset", cmdRset)
}

func cmdRset(c *Client, supervisor api.Supervisor, args string) {
	c.resetMessageState()

	sendResponse := true

	if c.Session != nil {
		if messageSended := sendedErrorToClientIfReject(c, c.Session.ResetMessageReceivingStatus()); messageSended {
			sendResponse = false
		}
	}

	if sendResponse {
		c.WriteCmd("250 OK")
	}

	return
}
