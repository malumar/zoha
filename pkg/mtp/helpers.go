package mtp

import (
	"github.com/google/uuid"
	"github.com/malumar/domaintools"
	"github.com/malumar/parsemail"
	"github.com/malumar/zoha/api"
	"github.com/malumar/zoha/pkg/mimemsg"
	"github.com/malumar/zoha/pkg/mtp/headers"
	"log/slog"
	"net/mail"
	"path"
	"strings"
	"time"
)

// It means that we have already sent a message from this email
// and if it is again send a copy to then in this case we ignore the subsequent sending of the copy
const XForwardedToHeader = "X-Forwarded-To"

// GetAbsoluteMaildirPath build maildir path prefix
// @defaultMailStorePath is default startpath prefix, if mb.MountPath is not empty it overwrite defaultMailStorePath
func GetAbsoluteMaildirPath(defaultMailStorePath string, mb *api.Mailbox) string {
	var sb strings.Builder
	if len(mb.MountPath) > 0 {
		sb.WriteString("/mnt/" + mb.MountPath)
	}

	// fixme! It's quite dangerous, but we have to have it for testing
	sb.WriteString(defaultMailStorePath)

	sb.WriteString(mb.Account)
	sb.WriteString("/domains/")
	sb.WriteString(mb.DomainLowerAscii)
	sb.WriteString("/mail/")
	sb.WriteString(mb.EmailLowerAscii)
	sb.WriteString("/mbox/")
	return sb.String()
}

func AllowAuthorize(service api.Service, mb *api.Mailbox, useSsl api.Maybe) bool {
	switch service {
	case api.SmtpInService:
		if !mb.SmtpInEnabled {
			slog.Error("AllowAuthorize: mailbox smtp in service disabled")
			return false
		}
		break
	case api.SmtpOutService:
		if !mb.SmtpOutEnabled {
			slog.Error("AllowAuthorize: mailbox smtp out service disabled")
			return false
		}
		break
	case api.ImapService:
		if !mb.ImapEnabled {
			slog.Error("AllowAuthorize: mailbox imap service disabled")
			return false
		}
		if mb.ImapUnsecureOff {
			if useSsl != api.Yes {
				slog.Error("AllowAuthorize: mailbox unsecure imap service disabled")
				return false
			}
		}
		break
	case api.PopService:
		if !mb.PopEnabled {
			slog.Error("AllowAuthorize: mailbox pop service disabled")
			return false
		}
		if mb.PopUnsecureOff {
			if useSsl != api.Yes {
				slog.Error("AllowAuthorize: mailbox pop imap service disabled")
				return false
			}
		}

		break
	default:
		slog.Error("AllowAuthorize: unknown service")
	}

	if mb.IsAlias {
		slog.Error("AllowAuthorize: this is alias, you can't login")
		return false
	}

	if mb.AccountDisabled {
		slog.Error("AllowAuthorize: this is account_disabled, you can't login")
		return false

	}
	if mb.Disabled {
		slog.Error("AllowAuthorize: this is disabled mailbox, you can't login")
		return false
	}

	return true
}

func SendAllRequiredCopies(logger *slog.Logger, fromMailerDaemon bool, originalRecipient string, mb *api.Mailbox,
	d *Delivery, proxy MessageReceiverProxy, supervisor api.MaildirSupervisor) {
	sendCopiesTo := GetSendCopyTo(logger, fromMailerDaemon, mb, d, proxy, supervisor)
	scl := len(sendCopiesTo)
	if scl == 0 {
		return
	}

	for no, toEmail := range sendCopiesTo {

		newDelivery := CloneDelivery(d, ToMailDir(toEmail))

		// Oznacza, ze wysłaliśmy to jako kopię z tego adresu
		// jeżeli w przyszlosci bedzie to wystepowalo to nie mozesz ponownie wyslac wiadomosci
		// do takiej osoby, poniewaz oznacza to ze bedzie zapetlenie
		// ..
		// .. to chyba niema znaczenia, bo zawsze trzeba przeszukać wszystkie DeliveredTo, jeśli występuje tam ten adres
		// to wychodzi na to samo
		//newDelivery.AddHeaderAtTop(headers.DeliveredTo, toEmail)
		//newDelivery.AddHeaderAs(XForwardedToHeader, toEmail, false)
		DeliveryAddHeaderAs(&newDelivery, XForwardedToHeader, toEmail, false)
		//		fmt.Println("GetHeader", newDelivery.HeaderValues(proxy.GetMessage(), headers.DeliveredTo))
		filename := supervisor.GenerateNextSpooledMessageName()

		DeliveryAddHeaderAs(&newDelivery, headers.XAction, originalRecipient+"|"+toEmail, true)

		_, err := StoreMessage(nil, proxy, newDelivery, filename)
		if err != nil {
			// well, it's hard
			logger.Error(
				"ProcessDeliver: message save error",
				"account", mb.Account,
				"from", mb.EmailLowerAscii,
				"dst", toEmail,
				"pos", no+1,
				"total", scl,
				"err", err.Error())
		} else {
			filenameInQueue := path.Base(filename)
			logger.Info("ProcessDeliver: message saved",
				"account", mb.Account,
				"from", mb.EmailLowerAscii,
				"dst", toEmail,
				"pos", no+1,
				"total", scl,
				"filenameInQueue", filenameInQueue,
				"filename", filename,
			)

		}

		newDelivery.Close()
	}

}
func GetSendCopyTo(logger *slog.Logger, fromMailerDaemon bool, mb *api.Mailbox, delivery *Delivery, proxy MessageReceiverProxy,
	supervisor api.Supervisor) (sendCopiesTo []string) {

	if !fromMailerDaemon && !delivery.HeaderContainsPattern(proxy.GetMessage(),
		[]string{"Precedence", "X-Autoresponder"}, "*") {
		sendCopiesTo = FilterSendCopyToEmails(logger, mb, supervisor,
			func(lowerAsciiEmailAddress string) bool {
				values := delivery.HeaderValues(proxy.GetMessage(), XForwardedToHeader)
				for _, val := range values {
					if val == mb.EmailLowerAscii {
						logger.Error("ProcessDelivery.Filter: message loop, skip",
							"account", mb.Account, "from", mb.EmailLowerAscii, "dst", lowerAsciiEmailAddress)
						return false

					}
					if val == lowerAsciiEmailAddress {
						logger.Error("ProcessDelivery.Filter: message already forwarded, skip",
							"account", mb.Account, "from", mb.EmailLowerAscii, "dst", lowerAsciiEmailAddress)
						return false
					}
				}

				return true
			})
	}

	return
}

// FilterSendCopyToEmails Filters email addresses Send-Copies to returns unique email addresses
// that you can send forward are supported by us or we've given you the green light to send to them from the machine
func FilterSendCopyToEmails(logger *slog.Logger, mb *api.Mailbox, supervisor api.Supervisor,
	allowToHandler func(lowerAsciiEmailAddress string) bool) []string {
	if mb == nil {
		logger.Error("mailbox is not initialized")
		return nil
	}

	uniq := make(map[string]bool)
	var count int
	transportTo := strings.TrimSpace(mb.TransportTo)
	if len(transportTo) == 0 {
		//		logger.Info("")
		return nil
	}
	lines := strings.Split(transportTo, ",")
	for _, line := range lines {
		dst := strings.ToLower(line)
		if !(domaintools.IsValidEmailAddress(dst)) {
			logger.Error("SendCopyToEmails not valid destination email, skip", "dst", dst,
				"from", mb.EmailLowerAscii, "account", mb.Account)
			continue
		}
		parts := strings.Split(dst, "@")
		if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) == 0 {
			logger.Error("SendCopyToEmails not valid destination email, skip", "dst", dst,
				"from", mb.EmailLowerAscii, "account", mb.Account)
			continue
		}

		// check if he is not trying to send to himself
		if dst == mb.EmailLowerAscii {
			logger.Error("SendCopyToEmails trying to send to himself (potential loop), skip", "dst", dst,
				"from", mb.EmailLowerAscii, "account", mb.Account)
			continue

		}
		if !allowToHandler(dst) {
			continue
		}
		if _, found := uniq[dst]; found {
			logger.Warn("SendCopyToEmails destination email already exist in list, skip", "dst", dst,
				"from", mb.EmailLowerAscii, "account", mb.Account)

			continue
		}

		if supervisor.IsLocalDomain(parts[1]) != api.Yes {
			logger.Warn("SendCopyToEmails destination email is not approved to sendy copy, skip", "dst", dst,
				"from", mb.EmailLowerAscii, "account", mb.Account)
			continue
		}

		uniq[dst] = true
		count++
	}

	if count == 0 {
		return nil
	}
	ret := make([]string, count)
	var i int

	for emial, _ := range uniq {
		ret[i] = emial
		i++
	}

	return ret

}

func SendAutoresponseMessage(logger *slog.Logger, msg *mail.Message, mb *api.Mailbox, d *Delivery) (string, *mimemsg.Message) {

	var subjectAr string
	toEmail := d.From.String()

	if !mb.AutoresponderEnabled {
		return "", nil
	}
	if d.From.IsEmpty() {
		logger.Error("autoresponder: can't send to empty FROM, skipping",
			"account", mb.Account, "from", mb.EmailLowerAscii)
		return "", nil
	}

	if strings.TrimSpace(strings.ToLower(toEmail)) == d.To.String() {
		logger.Error("autoresponder: can't send response to myself (potential loop), skipping",
			"account", mb.Account, "from", mb.EmailLowerAscii)
		return "", nil

	}
	if len(mb.AutoresponderBody) == 0 {
		logger.Error("autoresponder: response message don't have body, skipping",
			"account", mb.Account, "from", mb.EmailLowerAscii, "to", d.To.String())

		return "", nil
	}

	if d.HeaderContainsPattern(msg, []string{"X-Spam-Status"},
		"*SPF_SOFTFAIL*",
		"*RBL*", "*REDIR", "*ZGODA*",
		"*NESWLETTER*",
		"*SUBSCRIPTION_TEST*",
		"*DBL_SPAM*",
		"*URIBL_BLACK*",
	) ||
		d.HeaderContainsPattern(msg, []string{"Precedence"}, "*") {
		logger.Error("autoresponder: cant' respond to potential spam or bulk messages, skipping",
			"account", mb.Account, "from", mb.EmailLowerAscii, "to", d.To.String())

		return "", nil

	}

	if d.HeaderContainsPattern(msg, []string{"From", "Reply-To"},
		"*noreply@*", "*donotreply@*", "*FROM_MAILER*", "*FROM_DAEMON*", "*bounce*",
	) {
		logger.Error("autoresponder: cant' respond to noreply, skipping",
			"account", mb.Account, "from", mb.EmailLowerAscii, "to", d.To.String())

		return "", nil

	}

	if d.HeaderContainsPattern(msg, []string{"X-Mailer"},
		"*",
	) {
		logger.Error("autoresponder: cant' respond to mass message, skipping",
			"account", mb.Account, "from", mb.EmailLowerAscii, "to", d.To.String())
		return "", nil

	}

	if d.HeaderContainsPattern(msg, []string{
		"X-Autoresponder", "Auto-Submitted", "List-Id", "List-Unsubscribe",
		XForwardedToHeader,
	},
		"*",
	) {
		logger.Error("autoresponder: cant' respond to forwarded messages, skipping",
			"account", mb.Account, "from", mb.EmailLowerAscii, "to", d.To.String())
		return "", nil
	}

	if d.HeaderContainsPattern(msg, []string{"Received"},
		"*autorsponder*",
	) {
		logger.Error("autoresponder: cant' respond to forwarded messages (2), skipping",
			"account", mb.Account, "from", mb.EmailLowerAscii, "to", d.To.String())
		return "", nil

	}

	values := d.HeaderValues(msg, XForwardedToHeader)
	for _, val := range values {
		if strings.ToLower(val) == mb.EmailLowerAscii {
			logger.Error("autoresponder: cant' respond to messages already forwarded for me, skipping",
				"account", mb.Account, "from", mb.EmailLowerAscii, "to", d.To.String())

			return "", nil
		}
	}

	subjectAr = parsemail.DecodeMimeSentence(msg.Header.Get("Subject"))

	m := mimemsg.NewMessage()

	m.SetHeader("From", mb.EmailLowerAscii, mb.FullNameOfEmailOwner)
	m.SetHeader("To", toEmail)
	if len(strings.TrimSpace(mb.AutoresponderSubject)) == 0 {
		m.SetHeader("Subject", "RE: "+subjectAr)
	} else {
		m.SetHeader("Subject", strings.ReplaceAll(mb.AutoresponderSubject, "%t%", subjectAr))
	}

	m.SetHeader("Message-ID", uuid.New().String()+"@"+mb.DomainLowerAscii)
	m.SetHeader("MIME-Version", "1.0")
	m.SetHeader("X-Autoresponder", "true")
	m.SetHeader("Precedence", "bulk")
	m.SetDateHeader("Date", time.Now())

	m.SetBody("text/plain", mb.AutoresponderBody)

	return toEmail, m
}
