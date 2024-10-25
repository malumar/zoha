package data

import (
	"github.com/malumar/zoha/api"
	"github.com/malumar/zoha/pkg/mtp"
	"time"
)

// for testing purpose
const (
	ReverseHostname   = "serveros-lmtp-test"
	LmtpListenAddress = "0.0.0.0:2100"
	OfficeEmail       = "office@exmple.tld"
	MailerDaemon      = "MAILER_DAEMON@example.tld"
)

func NewData(hostname string, saltingPasswordHandler func(password string) string) []api.Mailbox {
	return []api.Mailbox{
		newMailbox(hostname, "superaccount", OfficeEmail, "The Example",
			saltingPasswordHandler("office_password"),
			"liam@example.tld,lisa@example.tld,wlilliam@example.tld"),
		newMailbox(hostname, "superaccount", "liam@exmple.tld", "Liam Smith",
			saltingPasswordHandler("liam_password"), ""),
		newMailbox(hostname, "superaccount", "lisa@exmple.tld", "Davie Jones",
			saltingPasswordHandler("davie_password"), ""),
		newMailbox(hostname, "superaccount", "wlilliam@exmple.tld", "Wlilliam Garcia",
			saltingPasswordHandler("wlilliam_password"), ""),
		newMailbox(hostname, "superaccount", "ceo@example.tld", "",
			"", "lisa@example.tld"),
	}
}

// newMailbox our simple database entry
func newMailbox(hostname, account, email, name, password string, transportTo string) api.Mailbox {
	isAlias := len(password) == 0
	e := mtp.SanitizeEmailAddress(email)
	return api.Mailbox{
		IsAlias:              isAlias,
		TransportTo:          transportTo,
		Created:              time.Now(),
		Account:              account,
		EmailLowerAscii:      e.String(),
		DomainLowerAscii:     e.Host,
		FullNameOfEmailOwner: name,
		Password:             password,
		// if this value will be empty,
		// message will be stored in the folder defined by Supervisor.defaultMailStorePath
		// MountPath:            filepath.Join(os.TempDir(), "zocha", account),
		SmtpOutEnabled:       true,
		SmtpInEnabled:        true,
		ImapEnabled:          true,
		ImapUnsecureOff:      false,
		PopEnabled:           true,
		PopUnsecureOff:       false,
		WebmailOff:           false,
		ImapSharedOff:        true,
		ImapGroup:            "",
		ImapSharedGroup:      "",
		WhiteList:            []string{"*@gmail.com", "???@hey.com"},
		BlackList:            []string{"*@yahoo.com"},
		Packet:               "SuperPacket",
		MoveToSpam:           true,
		SysUid:               0,
		SysGid:               0,
		Disabled:             false,
		AccountDisabled:      false,
		AutoresponderEnabled: true,
		AutoresponderSubject: "I am currently out of the office",
		AutoresponderBody:    "I am currently out of the office and probably out-of-my-mind drunk.  Enjoy your workweek",
		SmtpInHost:           "192.168.0.100",
		SmtpOutHost:          "192.168.0.101",
		SmtpOutIp:            "192.168.0.102",
		ImapIp:               "192.168.0.103",
		// must be same as hostname of supervisor, because the configuration base on all hosts is the same,
		// and the supervisor himself decides which mailboxes to support
		ImapHostname:          hostname,
		SmtpOutHostname:       "",
		ImapHost:              "",
		ImapStoreLmtp:         "",
		UnlimitedOutgoingSmtp: false,
		PrivatePortSfx:        "",
	}

}
