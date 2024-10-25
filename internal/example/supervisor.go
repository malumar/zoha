package example

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/malumar/zoha/api"
	"github.com/malumar/zoha/internal/example/data"
	"github.com/malumar/zoha/pkg/mtp"
	"github.com/malumar/zoha/pkg/spool"
	"log/slog"
)

type SessionFactoryHandler func(supervisor api.MaildirSupervisor, l *slog.Logger) mtp.Session

func NewDummyMaildirSupervisor(ctx context.Context, sessionFactoryHandler SessionFactoryHandler,
	defaultMailStorePath, spoolPath string) *SupervisorExample {
	return New(ctx, sessionFactoryHandler,
		data.NewData(data.ReverseHostname, SaltPassword),
		data.MailerDaemon, data.ReverseHostname, defaultMailStorePath, spoolPath)
}

func NewDummySupervisor(ctx context.Context, sessionFactoryHandler SessionFactoryHandler) *SupervisorExample {
	return NewSupervisor(ctx, sessionFactoryHandler,
		data.NewData(data.ReverseHostname, SaltPassword),
		data.MailerDaemon, data.ReverseHostname)
}

func NewSupervisor(ctx context.Context, sessionFactoryHandler SessionFactoryHandler, mailboxes []api.Mailbox,
	mailerDaemon string, hostname string) *SupervisorExample {
	return New(ctx, sessionFactoryHandler, mailboxes, mailerDaemon, hostname, "", "")
}

func New(ctx context.Context, sessionFactoryHandler SessionFactoryHandler, mailboxes []api.Mailbox,
	mailerDaemon string, hostname string, defaultMailStorePath, spoolPath string) *SupervisorExample {

	s := &SupervisorExample{
		hostname:              hostname,
		spool:                 spool.New(5, spoolPath, 0750),
		mailerDaemon:          mailerDaemon,
		defaultMailStorePath:  defaultMailStorePath,
		mailboxes:             mailboxes,
		sessionFactoryHandler: sessionFactoryHandler,
	}
	return s
}

type SupervisorExample struct {
	spool                 *spool.Spool
	sessionFactoryHandler SessionFactoryHandler
	mailerDaemon          string
	hostname              string
	defaultMailStorePath  string
	mailboxes             []api.Mailbox
}

func (self *SupervisorExample) AbsoluteSpoolPath() string {
	return self.defaultMailStorePath
}

func (self *SupervisorExample) MainSenderNode() string {
	// todo get value from NATS
	panic("not implemented")
}

func (self *SupervisorExample) GetAbsoluteMaildirPath(mb *api.Mailbox) string {
	return mtp.GetAbsoluteMaildirPath(self.defaultMailStorePath, mb)
}

func (self *SupervisorExample) GenerateNextSpooledMessageName() string {
	return self.spool.GenFilename(self.Hostname())
}

func (self *SupervisorExample) Authorization(username string, password string,
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

func (self *SupervisorExample) FindMailbox(name string) *api.Mailbox {
	for i, mb := range self.mailboxes {
		if mb.EmailLowerAscii == name {
			return &self.mailboxes[i]
		}
	}
	return nil
}
func (self *SupervisorExample) MailerDaemonEmailAddress() string {
	return self.mailerDaemon
}

func (self *SupervisorExample) Hostname() string {
	return self.hostname
}

func (self *SupervisorExample) IsLocalDomain(asciiDomainNameLowerCase string) api.Maybe {
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

func (self *SupervisorExample) IsLocalEmail(emailAsciiLowerCase string) api.Maybe {
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

func (self *SupervisorExample) OpenSession(ci mtp.ConnectionInfo, l *slog.Logger) (session mtp.Session, err error) {

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

	//LmtpSession := simplelmtpsession.NewLmtpSession(self, l)
	LmtpSession := self.sessionFactoryHandler(self, l)
	return LmtpSession, nil
}

func (self *SupervisorExample) CloseSession(session mtp.Session) {
	session.Close()
}

func SaltPassword(plainPassword string) string {
	h := sha256.New()
	h.Write([]byte("Secret,Salt,Value"))
	h.Write([]byte(plainPassword))
	bs := h.Sum(nil)
	return base64.StdEncoding.EncodeToString([]byte(bs))
}
