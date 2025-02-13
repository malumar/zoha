package mtp

import (
	"fmt"
	"github.com/malumar/zoha/api"
	"github.com/malumar/zoha/pkg/emailutil"
	"strings"
)

func init() {
	RegisterCommand("mail", cmdMail)
}

func cmdMail(c *Client, supervisor api.Supervisor, args string) {
	pz := strings.Index(strings.ToUpper(args), " FROM:")
	if pz == 0 {
		if c.DidHelo() /*&& c.allowNextMessage() */ {
			var mailFrom string
			if c.state.fromCommandPassed {
				// if c.state.from.IsValid() {
				c.WriteCmd("503  Bad sequence of Commands 1")

				c.logger.Error("Bad sequence of Commands 1",
					ActionNameField, "mailFrom",
					DecisionKey, Reject,
				)

				c.ErrorsCount++
				return
			} else {

				mailFrom = args[pz+6:]
				//  we have: <spamer@bad.man> BODY=8BITMIME ??
				//  logan.Info("extract from mailfrom: `%s`, oryginal:`%s`", mailFrom, args)

				var isEmptyEmail bool
				//goland:noinspection GrazieInspection
				if closeTagPos := strings.Index(mailFrom, ">"); closeTagPos > -1 && closeTagPos < len(mailFrom) {

					/*
						As mandated by RFC5321 Section 4.5.5, DSNs MUST be sent with a NULL Return-Path (or MAIL FROM) sender.
						The From: header MAILER-DAEMON@example.com is set by the receiving MTA, based on the NULL sender address.

					*/

					if openTagPos := strings.Index(mailFrom, "<"); openTagPos > -1 {
						// whether mailFrom[openTagPos+1:closeTagPos] point to the same position,
						// i.e. there is no value: DSNs MUST be sent with a NULL Return-Path (or MAIL FROM) sender.
						if openTagPos+1 == closeTagPos {
							// empty address, so we treat it as ours
							isEmptyEmail = true
						}
					}

					//goland:noinspection GrazieInspection
					if !isEmptyEmail {
						// We divide BODY=8BITMIME Val=Other
						// into
						// 8BITMIME arguments - it means that ASCII is 8-bit, without it it is 7-bit by default,
						// but who checks it anymore
						// SMTPUTF8 - that the data in the headers is written in SMTPUTF8 and not in 8 or 7 bits
						if _, err := parseArgs(strings.Split(mailFrom[closeTagPos+1:], " ")); err != nil {
							c.logger.Error("From header parsing error",
								"err", err,
								ActionNameField, "mailFrom",
								FromField, mailFrom[closeTagPos+1:],
								OriginalValue, args,
								DecisionKey, Reject,
							)
							c.WriteCmd(fmt.Sprintf("550 %s", err.Error()))
							c.ErrorsCount++
							// infoLog(c, "REJECT FROM: <%s>  REASON: bad email address format", mailFrom[pos+1:])
							return

						} else {
							mailFrom = mailFrom[:closeTagPos+1]
						}

					} else {
						// empty address email
					}

				}

				//var sender mtp.AddressEmail
				name, host, erre := emailutil.ExtractEmail(mailFrom)

				// we cannot reject messages as lmtp due to the wrong address if they have already arrived here,
				// i.e. they need to be accepted, you can receive the address from the null sender or root,
				// root@mailhostname - but you have to accept it without the domain!
				// October 17, 2023
				/*
					if erre != nil &&
						!c.Flags.HasFlag(mtp.ReceiveFromNullSender) &&
						!emailutil.IsEmptyErrEmail(erre) {

						c.LogInfo().Str(ActionNameField, "mailFrom").
							Str(FromField, mailFrom).
							Str(OriginalValue, args).
							Err(erre).
							Str(mtp.DecisionKey, mtp.Reject).
							Msg("Bad mail address format MAIL FROM")

						c.WriteCmd(fmt.Sprintf("550 %s", erre.Error()))
						c.ErrorsCount++

						return

					}
				*/
				if erre != nil &&
					!c.Flags.HasFlag(ReceiveFromNullSender) &&
					!emailutil.IsEmptyErrEmail(erre) {
					// added to pass even in case of errors
					if !c.Flags.HasFlag(LmtpService) {
						c.logger.Error("Bad mail address format MAIL FROM",
							ActionNameField, "mailFrom",
							OriginalValue, args,
							"err", erre,
							DecisionKey, Reject,
						)

						c.WriteCmd(fmt.Sprintf("550 %s", erre.Error()))
						c.ErrorsCount++

						return

					}

				}
				// if they are empty, it doesn't matter to us,
				// because at this stage we already accept an empty email
				sender := NewAddressEmail(name, host)

				var isLocalSender bool

				if !sender.IsEmpty() {

					if answer := supervisor.IsLocalEmail(sender.String()); !answer.HaveAnswer() {
						c.logger.Error("I can't get information about whether the email is a local address, I don't allow continuation",
							ActionNameField, "mailFrom",
							FromField, sender.String(),
							DecisionKey, Reject,
						)
						c.WriteCmd("451 4.7.1 Unable to complete command")
						c.ErrorsCount++
						return
					} else {
						isLocalSender = answer.ToBool()
					}

					if isLocalSender {
						if !c.Flags.HasFlag(ReceiveFromLocalUsers) {
							c.logger.Error("The server does not accept mail from local users",
								FromField, sender.String(),
								ActionNameField, "mailFrom",
								DecisionKey, Reject,
							)
							c.WriteCmd(fmt.Sprintf("553 5.7.1 <%s>: Sender address rejected, sever not for local users", sender))
							c.ErrorsCount++
						}
					} else {
						if !c.Flags.HasFlag(ReceiveFromRemoteUsers) {
							c.logger.Error("The server does not accept mail from users who are not logged in",
								FromField, sender.String(),
								ActionNameField, "mailFrom",
								DecisionKey, Reject,
							)
							c.WriteCmd(fmt.Sprintf("553 5.7.1 <%s>: Sender address rejected, sever not logged in (server only for local users)", sender))
							c.ErrorsCount++

						}

					}
					// if we have authorization, let's check if we can send from a local address, if we provide one
					if c.Flags.HasFlag(Authorization) {
						c.logger.Debug("Authorization support is enabled, so we will check whether the user has logged in")
						if !isCanSendAsLocalUserIfIsLocal(supervisor, sender, c, isLocalSender) {
							return
						}

					} else {
						c.logger.Debug("The server does not support authorization")
					}
				}

				if !isLocalSender {
					/*
						// SPF check
						spf_request := C.SPF_request_new(Spf_server)
						var spf_response *C.SPF_response_t

						C.SPF_request_set_ipv4_str(spf_request, C.CString(c.RemoteAddr))
						//C.SPF_request_set_helo_dom(spf_request, C.CString(c.Helo))
						C.SPF_request_set_env_from(spf_request, C.CString(sender))
						C.SPF_request_query_mailfrom(spf_request, &spf_response)
						c.MessageListener.AddHeaderAtTop(mailnt.PrivateHeader, "Received-SPF", strings.TrimLeft(C.GoString(C.SPF_response_get_received_spf(spf_response)), "Received-SPF: "))
						//fmt.Printf("%s\n%s\nWynik:%s\nHeader:%s\n", C.GoString(C.SPF_strresult(C.SPF_response_result(spf_response))), C.GoString(C.SPF_response_get_smtp_comment(spf_response)), C.GoString(C.SPF_response_get_header_comment(spf_response)), C.GoString(C.SPF_response_get_received_spf(spf_response)))
						C.SPF_response_free(spf_response)
						// sprawdzenie SPF
					*/
				}

				if err := c.Session.AcceptMessageFromEmail(sender); err != nil {
					if sended := sendedErrorToClientIfReject(c, err); !sended {

						c.logger.Error("do not AcceptMessageFromEmail",
							ActionNameField, "mailFrom",
							FromField, sender.String(),
							DecisionKey, Reject,
						)

						c.WriteCmd("451 4.7.1 Unable to complete command")
						c.ErrorsCount++

					}

					return
				} else {

					c.WriteCmd("250 OK")
				}
				if sender.IsEmpty() {
					if c.Flags.HasFlag(ReceiveFromNullSender) {
						sender = SanitizeEmailAddress(supervisor.MailerDaemonEmailAddress())
					}
				}
				c.state.from = sender
				// only now I know that we passed the police station
				c.state.fromCommandPassed = true

				//err := c.MessageListener.from(sender)
				// nowa implementacja
				/*
					erro := c.Listener.ClientListener.OnMailFrom(sender, c)
					if erro == nil {
						err := c.MessageListener.from(sender)
						if err != nil {
							//fmt.Println(err)
							c.WriteCmd(err.Error())
							c.ErrorsCount++
						} else {
							//c.AppendMessage.HasMailFrom = true
							//c.AppendMessage.MailFrom = args[:6]
							c.WriteCmd("250 OK")

						}
					} else {
						c.WriteCmd(erro.Error())
						c.ErrorsCount++
					}
				*/
				/*
					isLocalSender := isLocalEmailAddress(sender)
						if isLocalSender && !c.IsLoggedIn() {
							c.WriteCmd(fmt.Sprintf("553 5.7.1 <%s>: Sender address rejected: not logged in", sender))
							c.ErrorsCount++
						} else {
							if !isLocalSender {

							}

							err := c.MessageListener.from(fmt.Sprintf("%s@%s", name, host))
							if err != nil {
								//fmt.Println(err)
								c.WriteCmd(err.Error())
								c.ErrorsCount++
							} else {
								//c.AppendMessage.HasMailFrom = true
								//c.AppendMessage.MailFrom = args[:6]
								c.WriteCmd("250 OK")

							}

						}
				*/

			}
		}
	} else {
		c.UnknownCommand("", args)

	}
	return
}

func isCanSendAsLocalUserIfIsLocal(supervisor api.Supervisor, sender AddressEmail, c *Client, isLocalSender bool) (allowSend bool) {
	if isLocalSender {
		// If not logged in, or if logged in but writing from the same domain but different mailbox,
		// we return it
		if !c.Session.IsLoggedIn() || !isAllowSendAs(c, sender) {
			disallowSendAsResponse(c, sender)
			return false
		}
	}
	return true
}

// isAllowSendAs can I send from an email address
func isAllowSendAs(c *Client, sender AddressEmail) bool {
	allowSend, err := c.Session.IsAllowSendAs(sender)
	if err != nil {
		if sended := sendedErrorToClientIfReject(c, err); !sended {
			c.logger.Error("Shipment denied as specified user",
				ActionNameField, "mailFrom:IsAllowSendAs",
				FromField, sender.String(),
				DecisionKey, Reject,
			)
			c.WriteCmd("451 4.7.1 Unable to complete command")
			c.ErrorsCount++
		}
		return false
	}
	return allowSend
}

func disallowSendAsResponse(c *Client, sender AddressEmail) {
	c.WriteCmd(fmt.Sprintf("553 5.7.1 <%s>: Sender address rejected: not logged in", sender.String()))
	c.ErrorsCount++
}
