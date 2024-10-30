package mockup

import (
	"errors"
	"github.com/malumar/zoha/api"
	"github.com/malumar/zoha/pkg/mtp"
	"log/slog"
	"net/mail"
	"net/smtp"
	"os"
	"sync"
)

var LastCon *smtp.Client
var Wg sync.WaitGroup
var LmtpServer *mtp.Listener
var LmtpSession *SessionImpl

type SupervisorMockup struct{}

func (t *SupervisorMockup) Authorization(username string, password string, service api.Service, useSsl api.Maybe) *api.Mailbox {
	//TODO implement me
	panic("implement me")
}

func (t *SupervisorMockup) FindMailbox(name string) *api.Mailbox {
	//TODO implement me
	panic("implement me")
}

func (t *SupervisorMockup) GenerateNextSpooledMessageName() string {
	//TODO implement me
	panic("implement me")
}

func (t *SupervisorMockup) IsLocalEmail(emailAsciiLowerCase string) api.Maybe {
	//TODO implement me
	if IsLocalUser(emailAsciiLowerCase) {
		return api.Yes
	}
	return api.No
}

func (t *SupervisorMockup) MailerDaemonEmailAddress() string {
	return "MAILER_DAEMOB@supervisor-mockup-hostname"
}

func (t *SupervisorMockup) Hostname() string {
	return "supervisor-mockup-hostname"
}

func (t *SupervisorMockup) IsLocalDomain(asciiDomainNameLowerCase string) api.Maybe {
	for _, e := range LocalUsers {
		if e.Host == asciiDomainNameLowerCase {
			return api.Yes
		}
	}
	return api.No
}

func (t *SupervisorMockup) IsLocalAddressEmail(emailAsciiLowerCase mtp.AddressEmail) api.Maybe {
	for _, e := range LocalUsers {
		if e.IsEqual(emailAsciiLowerCase) {
			return api.Yes
		}
	}
	return api.No
}

func (t *SupervisorMockup) OpenSession(ci mtp.ConnectionInfo, l *slog.Logger) (session mtp.Session, err error) {
	LmtpSession = NewLmtpSession(&ci, l)
	return LmtpSession, nil
}

func (t *SupervisorMockup) CloseSession(session mtp.Session) {
	session.Close()
}

func NewLmtpSession(ci *mtp.ConnectionInfo, l *slog.Logger) *SessionImpl {
	return &SessionImpl{
		logger:                      l,
		username:                    LocalUser1Email,
		password:                    TestPassword,
		allowSendFrom:               []string{LocalUser1Email, LocalUser2Email},
		returnAuthErr:               errors.New("not implemented Auth"),
		returnNewMessageFromErr:     errors.New("not implemented New message from"),
		returnAddRecipientErr:       errors.New("not implemented Add Recipient"),
		returnOnReceivingMessageErr: errors.New("OnReceivingMessageErr"),

		loggedIn: false,
	}
}

type SessionImpl struct {
	logger                      *slog.Logger
	maxMessageSize              uint64
	username                    string
	password                    string
	allowSendFrom               []string
	recipients                  []mtp.AddressEmail
	from                        mtp.AddressEmail
	returnAuthErr               error
	returnNewMessageFromErr     error
	returnAddRecipientErr       error
	returnOnReceivingMessageErr error

	loggedIn bool
}

func (s SessionImpl) MaxMessageSize() uint64 {
	return s.maxMessageSize
}
func (s SessionImpl) IsLoggedIn() bool {
	return s.loggedIn
}

func (s SessionImpl) OnAuthorization(username string, password string, service api.Service, useSsl api.Maybe) (bool, error) {
	s.loggedIn = s.username == username && s.password == password

	return s.loggedIn, s.returnAuthErr
}

// IsAllowSendAs The server knows that you are logged in, the question is whether you can use this address as a sender
func (s SessionImpl) IsAllowSendAs(addressEmailInAsciiLowerCase mtp.AddressEmail) (bool, error) {
	for _, e := range s.allowSendFrom {
		if e == addressEmailInAsciiLowerCase.String() {
			return true, nil
		}
	}
	return false, nil
}

func (s SessionImpl) ResetMessageReceivingStatus() error {
	s.from = mtp.AddressEmail{}
	s.recipients = nil
	return nil
}

func (s SessionImpl) AcceptMessageFromEmail(senderAscii mtp.AddressEmail) error {

	switch senderAscii.String() {
	case SpamerEmail:
		return mtp.NewRejectErr(mtp.YouSpamerGoAwayRejectErr)
	case ThrotleEmial:
		return mtp.NewRejectErr(mtp.TooManyConnectionsRejectErr)
	default:
		/*
			for _, v := range LocalUsers {
				if v.IsEqual(senderAscii) {
					if !s.loggedIn {

					}
				}
			}*/
		return nil

	}

}

func (s SessionImpl) From() mtp.AddressEmail {
	return s.from
}

func (s SessionImpl) AcceptRecipient(recipientAscii mtp.AddressEmail) error {
	return nil
}

func (s *SessionImpl) Close() {
	s.logger = nil
	s.recipients = nil
	s.allowSendFrom = nil
	s.returnAddRecipientErr = nil
	s.returnOnReceivingMessageErr = nil
	s.returnAuthErr = nil
	s.returnNewMessageFromErr = nil

}

func (s *SessionImpl) AcceptMessage(message *mail.Message) error {
	// chec sie nam tu filtrowaC?
	// bez sensu odrazu do skrzynki
	return nil
}

func (s *SessionImpl) ProcessDelivery(proxy mtp.MessageReceiverProxy, delivery mtp.Delivery, hostname string) error {

	if _, _, err := mtp.DefaultNewFileFromDelivery(proxy, delivery, hostname, baseStoragePath); err != nil {
		return err
	}

	return nil
}

var baseStoragePath string

func init() {
	if p, err := os.MkdirTemp("", "zoha-temp*"); err != nil {
		panic(err)
	} else {
		baseStoragePath = p
	}
}
