package mtp

import (
	"fmt"
	"github.com/malumar/zoha/api"
	"github.com/malumar/zoha/pkg/emailutil"
	"strings"
)

func init() {
	RegisterCommand("rcpt", cmdRcpt)
}

func cmdRcpt(c *Client, supervisor api.Supervisor, args string) {
	pz := strings.Index(strings.ToUpper(args), " TO:")
	if pz == 0 {
		if c.DidHelo() && c.DidMailFrom() {

			name, host, erre := emailutil.ExtractEmail(args[pz+4:])
			if erre != nil {
				c.logger.Error("bad recipient email address format",
					ActionNameField, "rcptTo",
					DecisionKey, Reject,
					"err", erre,
					FromField, c.state.from.String(),
					ToField, args[pz+4:],
				)

				c.WriteCmd(fmt.Sprintf("501 %s", erre.Error()))
				c.ErrorsCount++
				return
				//infoLog(c, "REJECT FROM: <%s> TO: <%s> REASON: bad email address format", args[pz+4:])
			}

			rcpt := NewAddressEmail(name, host)
			//								if c.Listener.MaxRecipientsPerMessage > 0 && c.MessageListener.RecipientsCount() < c.Listener.MaxRecipientsPerMessage {
			if c.MaxRecipientsPerMessage > 0 && (c.state.receiver.RecipientsCount() >= c.MaxRecipientsPerMessage) {
				c.logger.Error("too many recipients of message",
					ActionNameField, "rcptTo",
					DecisionKey, Reject,
					DeliveryStatusField, DeliveryRejected,
					FromField, c.state.from.String(),
					ToField, rcpt.String(),
					LimitField, c.MaxRecipientsPerMessage,
					HasField, c.state.receiver.RecipientsCount(),
				)
				c.WriteCmd("452 Too many recipients\n221 Shutdown by errors")
				// just for the statistics
				c.ErrorsCount++
				c.WeClose = true

				return
			}
			//recipient := fmt.Sprintf("%s@%s", name, host)

			if !c.state.receiver.AcceptRecipient(supervisor.Hostname(), c.state.from, rcpt, c) {
				c.logger.Error("recipient address not accepted RCPT TO",
					ActionNameField, "rcptTo",
					DecisionKey, Reject,
					FromField, args[pz+4:],
					ToField, rcpt.String(),
				)
				return
			}

			c.WriteCmd(OkResponse)
			/*
				if err := c.Session.AcceptRecipient( mtp.NewAddressEmail(name, host)); err != nil {
					if sended := sendedErrorToClientIfReject(c, err); !sended {
						c.WriteCmd("431 4.7.1 Serwer odbiorcy nie odpowiada / The recipient’s server is not responding")
						c.WriteCmd("221 Closing connection")
						c.WeClose = true
						c.ErrorsCount++
					}
					return
				}
			*/
			/*
				recipient := fmt.Sprintf("%s@%s", name, host)
				// we check whether someone wants to send a message to another server without being logged in
				if !c.IsLoggedIn() && !isLocalEmailAddress(recipient) {
					c.WriteCmd(fmt.Sprintf("550 Requested action not taken: mailbox \"%s\" unavailable", recipient))
					c.ErrorsCount++
				} else {
					err := c.MessageListener.Recipient(recipient)
					if err != nil {
						c.WriteCmd(err.Error())
						c.ErrorsCount++
					} else {
						//c.AppendMessage.RecipientsCount++
						c.WriteCmd("250 OK")

					}

				}
			*/

		}

	} else {

		c.UnknownCommand("", args)
	}
	return
}
