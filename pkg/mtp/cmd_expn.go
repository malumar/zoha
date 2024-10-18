package mtp

import "github.com/malumar/zoha/api"

func init() {
	RegisterCommand("expn", cmdExpn)
}

func cmdExpn(c *Client, supervisor api.Supervisor, args string) {
	c.WriteCmd("504 Command not implemented")
	c.logger.Error("Command not implemented",
		ActionNameField, "cmdExpn",
		DecisionKey, Reject,
	)
}
