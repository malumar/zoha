package mtp

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/malumar/zoha/api"
	"strings"
)

func init() {
	RegisterCommand("auth", cmdAuth)
}

func cmdAuth(c *Client, supervisor api.Supervisor, args string) {
	if !c.Flags.HasFlag(Authorization) {
		_ = c.WriteCmd("504 Command not implemented")
		c.logger.Error("not implemented, authorization not enabled",
			ActionNameField, "cmdAuth",
			DecisionKey, Reject,
		)
		c.ErrorsCount++
		return
	}

	if len(args) < 2 {
		_ = c.WriteCmd("501 Syntax error in parameters or arguments")

		c.logger.Error("syntax error in parameters or arguments",
			ActionNameField, "cmdAuth",
			DecisionKey, Reject,
		)

		c.ErrorsCount++
		return
	}
	params := strings.Split(args[1:], " ")
	if cap(params) < 1 {
		_ = c.WriteCmd("501 Syntax error in parameters or arguments")
		c.ErrorsCount++
		c.logger.Error("Syntax error in parameters or arguments (2)",
			ActionNameField, "cmdAuth",
			DecisionKey, Reject,
		)

		return
	}
	authByType(params, c, args)

}

func authByType(params []string, c *Client, args string) {
	switch params[0] {
	case "PLAIN":
		if cap(params) == 2 {
			data, err := base64.StdEncoding.DecodeString(params[1])
			if err != nil || len(data) < 4 {
				_ = c.WriteCmd("501 Not Base64 string")
				c.logger.Error("Not Base64 string",
					ActionNameField, "cmdAuth",
					DecisionKey, Reject,
				)

				c.ErrorsCount++

			} else {

				authData := bytes.Split(data, []byte{0})
				if len(authData) == 3 && len(authData[0]) == 0 && len(authData[1]) > 0 && len(authData[2]) > 0 {

					_ = c.AuthMe(string(authData[1]), string(authData[2]))

				} else {
					_ = c.WriteCmd("501 Not Base64 password")
					c.logger.Error("Not Base64 password",
						ActionNameField, "cmdAuth",
						DecisionKey, Reject,
					)

					c.ErrorsCount++
				}

			}
		} else {
			_ = c.WriteCmd("501 Syntax go in parameters or arguments")
			c.logger.Error("Syntax go in parameters or arguments",
				ActionNameField, "cmdAuth",
				DecisionKey, Reject,
			)
			c.ErrorsCount++
		}
	case "LOGIN":
		c.state.callFunc = c.AuthLoginGetUsername
		_ = c.WriteCmd(fmt.Sprintf("334 %s", base64.StdEncoding.EncodeToString([]byte("Username:"))))

	default:
		var mechanism = strings.Trim(args, " ")
		_ = c.WriteCmd(fmt.Sprintf("504 The requested authentication mechanism \"%s\" is not supported", mechanism))
		c.logger.Error("The requested authentication mechanism is not supported",
			ActionNameField, "cmdAuth",
			DecisionKey, Reject,
			"mechanism", mechanism,
		)

		c.ErrorsCount++
	}
}
