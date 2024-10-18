package mtp

import (
	"fmt"
	"github.com/malumar/domaintools"
	"github.com/malumar/zoha/pkg/bitmask"
	"github.com/malumar/zoha/pkg/emailutil"
	"strings"
)

const (
	// 64 flags
	// Authorization support
	Authorization bitmask.Flag64 = 1 << iota // 1 << 0 which is 00000001
	StartTls                                 // 1 << 1 which is 00000010
	// Receive emails from local users
	ReceiveFromLocalUsers // 1 << 2 which is 00000100
	// Receive emails from remote (not logged in) users
	ReceiveFromRemoteUsers
	AllowLocalhostConnection
	// As mandated by RFC5321 Section 4.5.5, DSNs MUST be sent with a NULL Return-Path (or MAIL FROM) sender.
	// The From: header MAILER-DAEMON@example.com is set by the receiving MTA, based on the NULL sender address.
	// czyli domyślnie zezwalamy na puste maile i uwaga jeżeli je przyjmiesz nie możesz ich odrzucić
	ReceiveFromNullSender
	RequireLocalHostAuthorization
	// we ignore many errors, e.g. wrong sender, etc
	// then check the SMTP already
	LmtpService
)

type AuthorizationHandler func(username string, password string) (bool, error)

type Super interface {
	// Default hostname
	Hostname() string
	// Name visible when logging in
	Name() string
	// Software number (visible reporting processes)
	Version() string

	Flags() bitmask.Flag64

	// Used at the stage when we do not yet know or cannot determine the quota of the actual user
	DefaultMaxMessageSizeInBytes() int64
	DefaultMaxRecipientsPerMessage() int
}

type AddressEmail struct {
	Username string
	Host     string
}

func NewAddressEmail(username, host string) AddressEmail {
	return AddressEmail{
		Username: strings.ToLower(username),
		Host:     strings.ToLower(host),
	}
}

type PostmasterHolder interface {
	MailerDaemonEmailAddress() string
}

func NewPostmasterAddressEmail() {

}

func SanitizeEmailAddress(value string) AddressEmail {
	v, _ := NewAddressEmailSanitize(value)
	return v
}

func NewAddressEmailSanitize(value string) (AddressEmail, bool) {
	name, host, err := emailutil.ExtractEmail(domaintools.SafeAsciiDomainName(value))
	return AddressEmail{
		Username: name,
		Host:     host,
	}, err == nil
}

func (a AddressEmail) IsEmpty() bool {
	return len(a.Username) == 0 && len(a.Host) == 0
}

func (a AddressEmail) IsValid() bool {
	return a.Username != "" && a.Host != ""
}

func (a AddressEmail) IsEqual(another AddressEmail) bool {
	return a.String() == another.String()
}

func (e AddressEmail) String() string {
	if e.IsEmpty() {
		return ""
	}
	return e.Username + "@" + e.Host
}

func NewRejectErr(code RejectErrCode) error {
	return NewRejectErrMessage(code, "")
}

func NewRejectErrMessage(code RejectErrCode, message string) error {
	var close bool
	if val, ok := rejectMessages[code]; !ok {
		close = val.Close
	}
	return RejectErr{
		Code:            code,
		AppendMessage:   message,
		CloseConnection: close,
	}
}

type RejectErrCode int

const (

	// Reject should not be session-driven, should only inform about the following states:
	// 1 Server error
	TemporaryServerRejectErr RejectErrCode = iota

	// 2 Rejecting an email address (recipient or sender), it doesn't matter because it results from the context
	RecipientServerRejectErr

	// 3 Rejection due to message size
	// 4 Rejected due to lack of adequate free disk space
	QuotaLowRejectErr

	// 5 Spam rejection
	YouSpamerGoAwayRejectErr

	// 6 Rejection due to throtling
	TooManyConnectionsRejectErr

	// IN addition, it should be stated whether:
	// - temporarily or permanently (to inform whether to try further)
	// - be able to add their own message content

	// If the problem is temporary -- be sure to do your text
	TemporaryMailBoxProblemErr
	// Reject an attempt to send to the specified address (maybe you're a spammer,
	// or maybe the mailbox doesn't exist anymore) -- don't try again
	RecipientAddressRejectErr

	// we don't need more
)

type RejectBehaviour struct {
	Message string
	Close   bool
}

// Response code 4.7.1 means to try later,
// others should not return their own messages
// because the messages must be readable for Polish users the given messages are in Polish in the first place,
// it would have to be solved a little differently, e.g. the LocalTranslation array,
// which would be overwritten at the moment, can be overwritten messages in your own language
var rejectMessages = map[RejectErrCode]RejectBehaviour{
	QuotaLowRejectErr: {
		Message: "422 4.7.1 Skrzynka odbiorcy przekroczyla limit miejsca lub rozmiar wiadomosc przekracza limit rozmiaru odbiorcy dla poczty przychodzacej / The recipient’s mailbox is over its storage limit or size of the message exceeds the recipient’s size limits for incoming email",
		Close:   false,
	},
	TooManyConnectionsRejectErr: {
		Message: "450 4.2.1 Uzytkownik, z ktorym probujesz się skontaktować, odbiera poczte z szybkoscia uniemozliwiajaca dostarczenie dodatkowych wiadomosci / The user you are trying to contact is receiving mail at a rate that prevents additional messages from being delivered",
		Close:   false,
	},
	RecipientServerRejectErr: {
		Message: "431 4.7.1 Serwer odbiorcy nie odpowiada / The recipient’s server is not responding",
		Close:   false,
	},
	YouSpamerGoAwayRejectErr: {
		Message: "554 Czasowo wstrzymalismy odbior od ciebie poczty, ze względu na twoja zla reputacje / We have temporarily stopped receiving mail from you because of your bad reputation",
		Close:   true,
	},
	// moze byc tez wykorzystywany podczas odbioerania wiadomosci
	TemporaryServerRejectErr: {
		Message: "451 4.7.1 Unable to complete command",
		Close:   true,
	},
	TemporaryMailBoxProblemErr: {
		Message: "450 4.7.1 ",
		Close:   false,
	},
	RecipientAddressRejectErr: {
		Message: "541 ",
		Close:   false,
	},
}

type RejectErr struct {
	Code            RejectErrCode
	AppendMessage   string
	CloseConnection bool
}

func (r RejectErr) Error() string {
	return fmt.Sprintf("Odrzucenie: %s", RejectCodeToMessageWithMyMessage(r.Code, r.AppendMessage))
}

// ToMtpResponse To the replies in the servers (S|L)MTP
func (r RejectErr) ToMtpResponse() string {
	return RejectCodeToMessageWithMyMessage(r.Code, r.AppendMessage)
}

func RejectCodeToMessage(code RejectErrCode) string {
	return RejectCodeToMessageWithMyMessage(code, "")
}

func RejectCodeToMessageWithMyMessage(code RejectErrCode, appendMessage string) string {
	if val, ok := rejectMessages[code]; !ok {
		if appendMessage != "" {
			return fmt.Sprintf("%s %s (%d)", rejectMessages[TemporaryServerRejectErr].Message, appendMessage, int(code))
		} else {
			return fmt.Sprintf("%s (%d)", rejectMessages[TemporaryServerRejectErr].Message, int(code))
		}
	} else {
		if appendMessage != "" {

			return fmt.Sprintf("%s %s", val.Message, appendMessage)
		} else {
			return val.Message
		}

	}
}

func IsRejectErr(err error) bool {
	if _, ok := err.(RejectErr); ok {
		return true
	} else {
		return false
	}
}
