package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/malumar/zoha/api"
	"github.com/malumar/zoha/pkg/mtp"
	"github.com/malumar/zoha/pkg/spool"
	"log/slog"
	"time"
)

const officeEmail = "office@exmple.tld"

func NewSupervisor(hostname string, defaultMailStorePath, spoolPath string, ctx context.Context) *Supervisor {

	s := &Supervisor{
		hostname:             hostname,
		spool:                spool.New(5, spoolPath, 0750),
		mailerDaemon:         "MAILER_DAEMON@example.tld",
		defaultMailStorePath: defaultMailStorePath,
		mailboxes: []api.Mailbox{
			newMailbox(hostname, "superaccount", officeEmail, "The Example", "office_password", "liam@example.tld,lisa@example.tld,wlilliam@example.tld"),
			newMailbox(hostname, "superaccount", "liam@exmple.tld", "Liam Smith", "liam_password", ""),
			newMailbox(hostname, "superaccount", "lisa@exmple.tld", "Davie Jones", "davie_password", ""),
			newMailbox(hostname, "superaccount", "wlilliam@exmple.tld", "Wlilliam Garcia", "wlilliam_password", ""),
			newMailbox(hostname, "superaccount", "ceo@example.tld", "", "", "lisa@example.tld"),
		},
	}
	return s
}

type Supervisor struct {
	spool                *spool.Spool
	mailerDaemon         string
	hostname             string
	defaultMailStorePath string
	mailboxes            []api.Mailbox
}

func (self *Supervisor) GetAbsoluteMaildirPath(mb *api.Mailbox) string {
	return mtp.GetAbsoluteMaildirPath(self.defaultMailStorePath, mb)
}

func (self *Supervisor) GenerateNextSpooledMessageName() string {
	return self.spool.GenFilename(self.Hostname())
}

func (self *Supervisor) Authorization(username string, password string,
	service api.Service, useSsl api.Maybe) *api.Mailbox {

	if len(password) == 0 {
		slog.Error("Authorization: password is empty")
		return nil
	}

	mb := self.FindMailbox(username)
	if mb == nil {
		slog.Error("Authorization: username not found")
		return nil

	}

	if !mtp.AllowAuthorize(service, mb, useSsl) {
		return nil
	}

	if SaltPassword(password) != mb.Password {
		slog.Error("Authorization: this is wrong password, you can't login")
		return nil
	}

	return mb
}

func (self *Supervisor) FindMailbox(name string) *api.Mailbox {
	for i, mb := range self.mailboxes {
		if mb.EmailLowerAscii == name {
			return &self.mailboxes[i]
		}
	}
	return nil
}
func (self *Supervisor) MailerDaemonEmailAddress() string {
	return self.mailerDaemon
}

func (self *Supervisor) Hostname() string {
	return self.hostname
}

func (self *Supervisor) IsLocalDomain(asciiDomainNameLowerCase string) api.Maybe {
	var count int
	for _, v := range self.mailboxes {
		if v.DomainLowerAscii == asciiDomainNameLowerCase {
			if v.AccountDisabled {
				return api.DontKnow
			}
			if v.Disabled {
				continue
			}
			count++

		}
	}
	if count > 0 {
		return api.Yes
	}
	return api.No
}

func (self *Supervisor) IsLocalEmail(emailAsciiLowerCase string) api.Maybe {
	for _, v := range self.mailboxes {
		if v.EmailLowerAscii == emailAsciiLowerCase {
			if !v.AccountDisabled && !v.Disabled {
				return api.Yes
			}
			return api.DontKnow

		}
	}
	return api.Yes
}

var allowedIp = []string{"192.168.0.1", "127.0.0.1"}

func (self *Supervisor) OpenSession(ci mtp.ConnectionInfo, l *slog.Logger) (session mtp.Session, err error) {

	allow := false

	// Remember, as LMTP do not use implicit authorization
	for _, ip := range allowedIp {
		if ci.RemoteAddr == ip {
			allow = true
			break
		}
	}

	if !allow {
		if ci.ReverseHostName == "allowedHostname" {
			slog.Error(fmt.Sprintf("Hostname `%s` is not authorized", ci.ReverseHostName))
		}
		return nil, mtp.NewRejectErrMessage(mtp.TemporaryMailBoxProblemErr, "i don't know you")
	}

	LmtpSession := NewLmtpSession(self, l)
	return LmtpSession, nil
}

func (self Supervisor) CloseSession(session mtp.Session) {
	session.Close()
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
		Password:             SaltPassword(password),
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
func SaltPassword(plainPassword string) string {
	h := sha256.New()
	h.Write([]byte("Secret,Salt,Value"))
	h.Write([]byte(plainPassword))
	bs := h.Sum(nil)
	return base64.StdEncoding.EncodeToString([]byte(bs))
}
