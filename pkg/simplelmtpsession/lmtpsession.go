// Package simplelmtpsession very simple implementation of mtp.Session.
// rather for testing purposes
package simplelmtpsession

import (
	"errors"
	"fmt"
	"github.com/malumar/zoha/api"
	"github.com/malumar/zoha/pkg/bytefmt"
	"github.com/malumar/zoha/pkg/mimemsg"
	"github.com/malumar/zoha/pkg/mtp"
	"github.com/malumar/zoha/pkg/mtp/headers"
	"log/slog"
	"net/mail"
	"strings"
)

const (
	WhitelistedTag = "WL"
	// todo: BlacklistedTag = "BL"
	SpamTag      = "S"
	SocialTag    = "SC"
	AdvertiseTag = "AD"
)

type SenderEmailClassification int

// MaxSingelMessageSize max message size in megabytes
const MaxSingelMessageSize uint64 = 25 * (1024 * 1024)
const (
	SenderEmailNotClassified = iota
	SenderEmailIsSpam
	SenderEmailIsWhitelist
	SenderEmailIsSocial
	SenderEmailIsAdvertiser
)

var errAuthNotImplemented = errors.New("not implemented Auth")

var errNewMessageFromNotImplemented = errors.New("not implemented New message from")
var errAddRecipientNotImplemented = errors.New("not implemented Add Recipient")
var errOnReceivingMessage = errors.New("OnReceivingMessageErr")

func NewLmtpSession(supervisor api.MaildirSupervisor, l *slog.Logger) *SessionImpl {
	return &SessionImpl{
		logger:                      l,
		supervisor:                  supervisor,
		returnAuthErr:               errAuthNotImplemented,
		returnNewMessageFromErr:     errNewMessageFromNotImplemented,
		returnAddRecipientErr:       errAddRecipientNotImplemented,
		returnOnReceivingMessageErr: errors.New("OnReceivingMessageErr"),
		loggedIn:                    false,
		//hi:                          hi,
		//hostname:                    myHostname,
		//	outgoingSpool: nil,
	}
}

type SessionImpl struct {
	logger                      *slog.Logger
	supervisor                  api.MaildirSupervisor
	maxMessageSize              uint64
	allowSendFrom               []string
	senderClassification        SenderEmailClassification
	recipients                  []mtp.AddressEmail
	recipientsMb                []*api.Mailbox
	from                        mtp.AddressEmail
	returnAuthErr               error
	returnNewMessageFromErr     error
	returnAddRecipientErr       error
	returnOnReceivingMessageErr error
	loggedIn                    bool
}

func (self *SessionImpl) MaxMessageSize() uint64 {
	return MaxSingelMessageSize
}
func (self *SessionImpl) IsLoggedIn() bool {
	return self.loggedIn
}

func (self *SessionImpl) OnAuthorization(username string, password string, service api.Service, useSsl api.Maybe) (bool, error) {

	self.loggedIn = false

	mb := self.supervisor.Authorization(username, password, service, useSsl)
	if mb == nil {
		return false, self.returnAuthErr
	}
	self.loggedIn = true

	// If you have enabled the possibility of sending messages from this mailbox,
	// add an email LmtpListenAddress to the allowed senders
	if mb.SmtpOutEnabled {
		self.allowSendFrom = append(self.allowSendFrom, mb.EmailLowerAscii)
	}

	return true, nil
}

// IsAllowSendAs The server knows that you are logged in, the question is whether you can use this LmtpListenAddress as a sender
func (self *SessionImpl) IsAllowSendAs(addressEmailInAsciiLowerCase mtp.AddressEmail) (bool, error) {
	if self.loggedIn {
		from := addressEmailInAsciiLowerCase.String()
		for _, e := range self.allowSendFrom {
			if e == from {
				return true, nil
			}
		}
	}
	return false, nil
}

func (self *SessionImpl) ResetMessageReceivingStatus() error {
	self.from = mtp.AddressEmail{}
	self.recipients = nil
	self.clearRecipients()

	return nil
}

func (self *SessionImpl) From() mtp.AddressEmail {
	return self.from
}
func (self *SessionImpl) AcceptMessageFromEmail(senderAscii mtp.AddressEmail) error {
	// note senderAscii can be an empty LmtpListenAddress, so it will mean that it is an LmtpListenAddress to mailer_daemonmailer_filter
	// from your host and you need to add your host at the end or treat it as empty
	if senderAscii.IsEmpty() {
		//
	}

	switch senderAscii.String() {
	case "some-email-that-we-know-is-a-spammer":
		return mtp.NewRejectErr(mtp.YouSpamerGoAwayRejectErr)
	case "email-LmtpListenAddress-that-sends-to-us-too often":
		return mtp.NewRejectErr(mtp.TooManyConnectionsRejectErr)
	}

	// we accept so we don't return the error
	return nil
}

func (self *SessionImpl) AcceptRecipient(recipientAscii mtp.AddressEmail) error {

	re := recipientAscii.String()
	mb := self.supervisor.FindMailbox(re)
	if mb == nil {
		return fmt.Errorf("login %s not found", re)
	}

	if mb.ImapHostname != self.supervisor.Hostname() {
		//if mb.ImapHostname != self.hostname {
		self.logger.Info(fmt.Sprintf("Email assigned to '%self' and we handle '%self'",
			mb.ImapHostname, self.supervisor.Hostname()))
		return fmt.Errorf("login %self cannot receive messages at this time, "+
			"is assigned to another host / Uzytkownik %self nie moze w tej chwili odbierac wiadomosci,"+
			" sprobuj pozniej, jest przypisany do innego hosta", re, re)
	}

	if !mb.SmtpInEnabled {
		return fmt.Errorf("login %self cannot receive messages at this time / "+
			"Uzytkownik %self nie moze w tej chwili odbierac wiadomosci, sprobuj pozniej", re, re)
	}

	self.recipientsMb = append(self.recipientsMb, mb)
	return nil
}

func (self *SessionImpl) Close() {
	self.logger = nil
	self.supervisor = nil
	self.senderClassification = SenderEmailNotClassified
	self.recipients = nil
	self.returnAddRecipientErr = nil
	self.returnOnReceivingMessageErr = nil
	self.returnAuthErr = nil
	self.returnNewMessageFromErr = nil
	// self.allowSendFrom = nil
	self.clearRecipients()
}

func (self *SessionImpl) clearRecipients() {
	self.recipients = nil

	for i := range self.recipientsMb {
		self.recipientsMb[i] = nil
	}
	self.recipientsMb = nil
}

func (self *SessionImpl) AcceptMessage(message *mail.Message) error {
	// want to filter here? pointless immediately to the mailbox
	// todo: place your filters here
	return nil
}

func (self *SessionImpl) ProcessDeliveryProcessDelivery(proxy mtp.MessageReceiverProxy, delivery mtp.Delivery,
	reverseHostname string) error {
	// 4 stars or more
	spamCount := len(proxy.GetMessage().Header.Get("X-Spam-Level"))

	mb := self.getMailboxFromDelivery(delivery)
	if mb == nil {
		return fmt.Errorf("mailbox %s not found", delivery.To)
	}

	sender := delivery.From.String()
	senderClassification := SenderEmailNotClassified

	fromMailerDaemon := false
	if strings.ToLower(delivery.From.String()) == strings.ToLower(self.supervisor.MailerDaemonEmailAddress()) {
		fromMailerDaemon = true
	}

	if !fromMailerDaemon {

		if mtp.IsWhitelisted(mb, sender) {
			senderClassification = SenderEmailIsWhitelist

		} else {
			if mtp.IsBlacklisted(mb, sender) {
				senderClassification = SenderEmailIsSpam
			} else {
				if strings.HasSuffix(sender, "@facebookmail.com") {
					senderClassification = SenderEmailIsSocial

				}
			}
			if spamCount >= 3 {
				senderClassification = SenderEmailIsSpam

			}
		}

	}

	// put your message in maildir new
	delivery.IsRecent = true
	var tags []string
	// courier note: remember that if the user does not subscribe to the folder, he will not see these messages
	switch senderClassification {
	case SenderEmailIsSpam:
		tags = append(tags, SpamTag)
		// FIXME: only to the spam folder when the customer wants
		if mb.MoveToSpam {
			// FIXME if it contains RCVD_IN_MSPIKE_WL it means that the sender is OK and if you
			// 		 have doubts mark it as a good sender
			// FIXME whitelist can only be known if it is VALID DKIM or SPF without it you should not believe
			// 		 just based on the sender's email address because it can be fabricated
			delivery.Mailbox = ".SPAM"
		}
		break
	case SenderEmailIsSocial:
		delivery.Mailbox = ".SocialMedia"
		tags = append(tags, SocialTag)
		break
	case SenderEmailIsAdvertiser:
		delivery.Mailbox = ".Ads"
		tags = append(tags, AdvertiseTag)
		break
	case SenderEmailIsWhitelist:
		tags = append(tags, WhitelistedTag)
		break
	case SenderEmailNotClassified:
		break
	default:
		self.logger.Debug("not a classified sender", "to", delivery.To, "from", delivery.From)
		break
	}

	if len(tags) > 0 {
		var sb strings.Builder
		for _, k := range tags {
			sb.WriteRune('[')
			sb.WriteString(k)
			sb.WriteRune(']')
		}
		mtp.DeliveryAddHeaderAs(&delivery, headers.XServerosTag, sb.String(), false)
	}

	/*Received: from mail-pa0-f66.google.com (mail-pa0-f66.google.com [xx.xx.xxx.xx])
	by xxx.xxxx.xx (The Mail Program Name) with ESMTPS id 06CC31A6161A
	for <xxxxx@yyyyyy.zz>; Fri, 23 Aug 2013 16:43:12 +0200 (CEST)*/

	originalRecipient := delivery.To.String()

	var arMsg *mimemsg.Message
	var arToEmail string
	if mb.AutoresponderEnabled {
		if !fromMailerDaemon && (senderClassification == SenderEmailNotClassified || senderClassification == SenderEmailIsWhitelist) {
			arToEmail, arMsg = mtp.SendAutoresponseMessage(self.logger, proxy.GetMessage(), mb, &delivery)
		}
	}

	fn, ds, err := mtp.DefaultNewFileFromDelivery(proxy, delivery,
		self.supervisor.Hostname(),
		self.supervisor.GetAbsoluteMaildirPath(mb),
	)
	if err != nil {
		self.logger.Debug("failed to save incoming message")
		return err
	}

	var dsc uint64
	if ds > 0 {
		dsc = uint64(ds)
	}
	self.logger.Debug("saved file as `%s` size %s\n", fn, bytefmt.ByteSize(dsc))

	if arMsg != nil && len(arToEmail) > 0 {
		if _, err := mtp.StoreMessageFromWriter(
			[]byte(headers.XAction+": "+originalRecipient+"|"+arToEmail+"\n"),
			arMsg,
			self.supervisor.GenerateNextSpooledMessageName(),
		); err != nil {
			self.logger.Error("USER: %s EMAIL: %s AUTORESPONDER: %s save error %v",
				mb.Account,
				mb.EmailLowerAscii,
				arToEmail,
				err.Error())
		}
	}
	mtp.SendAllRequiredCopies(self.logger, fromMailerDaemon, originalRecipient, mb, &delivery, proxy, self.supervisor)
	return nil
}

func (self *SessionImpl) getMailboxFromDelivery(delivery mtp.Delivery) *api.Mailbox {
	find := delivery.To.String()
	for _, mb := range self.recipientsMb {
		if mb.EmailLowerAscii == find {
			return mb
		}
	}
	return nil
}
