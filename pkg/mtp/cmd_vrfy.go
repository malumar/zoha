package mtp

import "github.com/malumar/zoha/api"

func init() {
	RegisterCommand("vrfy", cmdVrfy)
}

func cmdVrfy(c *Client, supervisor api.Supervisor, args string) {
	c.logger.Error("Command not implemented",
		ActionNameField, "cmdVrfy",
		DecisionKey, Reject,
	)
	c.WriteCmd("504 Command not implemented")
	return
}
