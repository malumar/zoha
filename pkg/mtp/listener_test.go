package mtp

import (
	"context"
	"errors"
	"fmt"
	"github.com/malumar/zoha/api"
	"github.com/malumar/zoha/pkg/bitmask"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"net/mail"
	"net/smtp"
	"os"
	"sync"
	"testing"
)

const (
	TestServerAddress = ":38000"

	LocalUser1Email = "emma@example.tld"
	LocalUser2Email = "james@example.tld"
	LocalUser3Email = "ava@example.tld"
	LocalUser4Email = "fax@example.tld"

	LocalAlias1Src = "emma.smith@example.tld"
	LocalAlias1Dst = LocalUser1Email

	LocalAlias2Src = "james.brown@example.tld"
	LocalAlias2Dst = LocalUser2Email

	TestPassword = `~CJS/CA"KI`
	SpamerEmail  = "spamer@ugly.spammer"
	ThrotleEmial = "toofastw@broken.email-client"
	RemoteUser1  = "janek@company.any-domain"

	ServerHostname           = "server-hostname-test"
	MailerDaemonAddressEmail = "MAILER_DAEMON@" + ServerHostname
)

const message1 = `Return-Path: <mlemos@acm.org>
To: Manuel Lemos <mlemos@linux.local>
Subject: Testing Manuel Lemos' MIME E-mail composing and sending PHP class: HTML message
From: mlemos <mlemos@acm.org>
Reply-To: mlemos <mlemos@acm.org>
Sender: mlemos@acm.org
X-Mailer: http://www.phpclasses.org/mimemessage $Revision: 1.63 $ (mail)
MIME-Version: 1.0
Content-Type: multipart/mixed; boundary="652b8c4dcb00cdcdda1e16af36781caf"
Message-ID: <20050430192829.0489.mlemos@acm.org>
Date: Sat, 30 Apr 2005 19:28:29 -0300


--652b8c4dcb00cdcdda1e16af36781caf
Content-Type: multipart/related; boundary="6a82fb459dcaacd40ab3404529e808dc"


--6a82fb459dcaacd40ab3404529e808dc
Content-Type: multipart/alternative; boundary="69c1683a3ee16ef7cf16edd700694a2f"


--69c1683a3ee16ef7cf16edd700694a2f
Content-Type: text/plain; charset=ISO-8859-1
Content-Transfer-Encoding: quoted-printable

This is an HTML message. Please use an HTML capable mail program to read
this message.

--69c1683a3ee16ef7cf16edd700694a2f
Content-Type: text/html; charset=ISO-8859-1
Content-Transfer-Encoding: quoted-printable

<html>
<head>
<title>Testing Manuel Lemos' MIME E-mail composing and sending PHP class: H=
TML message</title>
<style type=3D"text/css"><!--
body { color: black ; font-family: arial, helvetica, sans-serif ; backgroun=
d-color: #A3C5CC }
A:link, A:visited, A:active { text-decoration: underline }
--></style>
</head>
<body>
<table background=3D"cid:4c837ed463ad29c820668e835a270e8a.gif" width=3D"100=
%">
<tr>
<td>
<center><h1>Testing Manuel Lemos' MIME E-mail composing and sending PHP cla=
ss: HTML message</h1></center>
<hr>
<P>Hello Manuel,<br><br>
This message is just to let you know that the <a href=3D"http://www.phpclas=
ses.org/mimemessage">MIME E-mail message composing and sending PHP class</a=
> is working as expected.<br><br>
<center><h2>Here is an image embedded in a message as a separate part:</h2>=
</center>
<center><img src=3D"cid:ae0357e57f04b8347f7621662cb63855.gif"></center>Than=
k you,<br>
mlemos</p>
</td>
</tr>
</table>
</body>
</html>
--69c1683a3ee16ef7cf16edd700694a2f--

--6a82fb459dcaacd40ab3404529e808dc
Content-Type: image/gif; name="logo.gif"
Content-Transfer-Encoding: base64
Content-Disposition: inline; filename="logo.gif"
Content-ID: <ae0357e57f04b8347f7621662cb63855.gif>

R0lGODlhlgAjAPMJAAAAAAAA/y8vLz8/P19fX19f339/f4+Pj4+Pz7+/v///////////////////
/////yH5BAEAAAkALAAAAACWACMAQwT+MMlJq7046827/2AoHYChGAChAkBylgKgKClFyEl6xDMg
qLFBj3C5uXKplVAxIOxkA8BhdFCpDlMK1urMTrZWbAV8tVS5YsxtxmZHBVOSCcW9zaXyNhslVcto
RBp5NQYxLAYGLi8oSwoJBlE+BiSNj5E/PDQsmy4pAJWQLAKJY5+hXhZ2dDYldFWtNSFPiXssXnZR
k5+1pjpBiDMJUXG/Jo7DI4eKfMSmxsJ9GAUB1NXW19jZ2tvc3d7f4OHi2AgZN5vom1kk6F7s6u/p
m3Ab7AOIiCxOyZuBIv8AOeTJIaYQjiR/kKTr5GQNE3pYSjCJ9mUXClRUsLxaZGciC0X+OlpoOuQo
ZKdNJnIoKfnxRUQh6FLG0iLxIoYnJd0JEKISJyAQDodp3EUDC48oDnUY7HFI3wEDRjzycQJVZCQT
Ol7NK+G0qgtkAcOKHUu2rNmzYTVqRMt2bB49bHompSchqg6HcGeANSMxr8sEa2y2HexnSEUTuWri
SSbkYh7BgGVAnhB1b2REibESYaRoBgqIMYx59tFM9AvQffVG49P5NMZkMlHKhJPJb0knmSKZ6kSX
JtbeF3Am7ocok6c7cM7pU5xcXiJJETUz16qPrzEfaFgZpvzn7h86YV5r/1mxXeAUMVyEIpnVUGpN
RlG2ka9b3lP3pm2l6u7P+l/YLj3+RlEHbz1C0kRxSITQaAcilVBMEzmkkEQO8oSOBNg9SN+AX6hV
z1pjgJiAhwCRsY8ZIp6xj1ruqCgeGeKNGEZwLnIwzTg45qjjjjz2GEA5hAUp5JBEFmnkkSCoWEcZ
X8yohZNK1pFGPQS4hx0qNSLJlk9wCQORYu5QiMd7bUzGVyNlRiOHSlpuKdGEItHQ3HZ18beRRyws
YSY/waDTiHf/tWlWUBAJiMJ1/Z0XXU7N0FnREpKM4NChCgbyRDq9XYpOplaKopN9NMkDnBbG+UMC
QwLWIeaiglES6AjGARcPHCWoVAiatcTnGTABZoLPaPG1phccPv366mEvWEFSLnj+2QaonECwcJt/
e1Zw3lJvVMmftBdVNQS3UngLCA85YHIQOy6JO9N4eZW7KJwtOUZmGwOMWqejwVW6RQzaikRHX3yI
osKhDAq8wmnKSmdMwNidSOof9ZG2DoV0RfTVmLFtGmNk+CoZna0HQnPHS3AhRbIeDpqmR09E0bsu
soeaw994z+rwQVInvqLenBftYjLOVphLFHhV9qsnez8AEUbQRgO737AxChjmyANxuEFHSGi7hFCV
4jxLst2N8sRJYU+SHiAKjlmCgz2IffbLI5aaQR71hnkxq1ZfHSfKata6YDCJDMAQwY7wOgzhjxgj
VFQnKB5uX4mr9qJ79pann+VcfcSzsSCd2mw5scqRRvlQ6TgcUelYhu75iPE4JejrsJOFQAG01277
7bjnrvvuvPfu++/ABy887hfc6OPxyCevPDdAVoDA89BHL/301Fdv/fXYZ6/99tx3Pz0FEQAAOw==

--6a82fb459dcaacd40ab3404529e808dc
Content-Type: image/gif; name="background.gif"
Content-Transfer-Encoding: base64
Content-Disposition: inline; filename="background.gif"
Content-ID: <4c837ed463ad29c820668e835a270e8a.gif>

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
eZmYmZmauZmc2ZlCEQEAOw==

--6a82fb459dcaacd40ab3404529e808dc--

--652b8c4dcb00cdcdda1e16af36781caf
Content-Type: text/plain; name="attachment.txt"
Content-Transfer-Encoding: base64
Content-Disposition: attachment; filename="attachment.txt"

VGhpcyBpcyBqdXN0IGEgcGxhaW4gdGV4dCBhdHRhY2htZW50IGZpbGUgbmFtZWQgYXR0YWNobWVu
dC50eHQgLg==

--652b8c4dcb00cdcdda1e16af36781caf--
`

type combinationHandler func(t *testing.T, f bitmask.Flag64, combinationName string)

var localMailbox = []AddressEmail{
	SanitizeEmailAddress(LocalUser1Email),
	SanitizeEmailAddress(LocalUser2Email),
	SanitizeEmailAddress(LocalUser3Email),
	SanitizeEmailAddress(LocalUser4Email),
}

var localAlias = []AddressEmail{
	SanitizeEmailAddress(LocalAlias1Src),
	SanitizeEmailAddress(LocalAlias2Src),
}

var localUsers = append(localMailbox, localAlias...)

func TestReadme(t *testing.T) {
	// use ramdisk for tests:
	// macos: diskutil erasevolume HFS+ "RAMDisk" `hdiutil attach -nomount ram://2048`
	//
	//256 MB = 524288
	//512 MB = 1048576
	//1 GB = 2097152
	//2 GB = 4194304
	//`hdiutil attach -nomount ram://2048`
	t.Log("Start the ramdisk hdiutil attach -nomount ram://2048 ")
}

func BenchmarkSending(b *testing.B) {

	for i := 0; i < b.N; i++ {

		lmtpServer = NewDefaultListenerExt(TestServerAddress, ServerHostname, ReceiveFromRemoteUsers|Authorization)
		ctx, cFunc := context.WithCancel(context.Background())

		if err := lmtpServer.Listen(); err != nil {

			return
		}

		wg.Add(1)
		go func() {
			wg.Done()
			lmtpServer.Run(ctx, NewTestSuperVisor(ServerHostname), cFunc)

		}()

		ctx.Done()
		wg.Wait()

		// assert.NoError(t,con.Hello("abcd"))
	}
}

func TestBase(t *testing.T) {
	runAllCombinations(t, func(t *testing.T, f bitmask.Flag64, combinationName string) {
		if f.HasFlag(ReceiveFromRemoteUsers) {
			assert.EqualError(t, con.Mail(SpamerEmail), RejectCodeToMessage(YouSpamerGoAwayRejectErr), combinationName)
			assert.EqualError(t, redial(t).Mail(ThrotleEmial), RejectCodeToMessage(TooManyConnectionsRejectErr), combinationName)
		} else {

			assert.EqualError(t, con.Mail(SpamerEmail), s("553 5.7.1 <%s>: Sender address rejected, sever not logged in (server only for local users)", SpamerEmail), combinationName)
			assert.EqualError(t, con.Mail(ThrotleEmial), s(rejectMessages[YouSpamerGoAwayRejectErr].Message), combinationName)
		}

		eqErrFromString(t,
			redial(t).Mail(LocalUser1Email),
			stringIf(
				f.HasFlag(ReceiveFromLocalUsers),
				stringIf(
					f.HasFlag(Authorization),
					s("553 5.7.1 <%s>: Sender address rejected: not logged in", LocalUser1Email),
					"",
				),
				fmt.Sprintf("553 5.7.1 <%s>: Sender address rejected, sever not for local users", LocalUser1Email),
			),
			combinationName+" Sending an email from a local user's address to the same address",
		)

		eqErrFromString(t,
			redial(t).Mail(LocalUser2Email),
			stringIf(f.HasFlag(ReceiveFromLocalUsers),
				stringIf(
					f.HasFlag(Authorization),
					fmt.Sprintf("553 5.7.1 <%s>: Sender address rejected: not logged in", LocalUser2Email),
					"",
				),
				s("553 5.7.1 <%s>: Sender address rejected, sever not for local users", LocalUser2Email),
			),
			combinationName+"Sending an email from the address of a selected local user to the address of another local user",
		)

		eqErrFromString(t,
			redial(t).Mail(RemoteUser1),
			stringIf(
				f.HasFlag(ReceiveFromRemoteUsers),
				"",
				fmt.Sprintf("553 5.7.1 <%s>: Sender address rejected, sever not logged in (server only for local users)", RemoteUser1),
			),
			combinationName+"Sending emails from the remote user's address to the local user's address",
		)
		return

	})
}

func TestServer(t *testing.T) {

	combination(t, ReceiveFromRemoteUsers|Authorization, "sendmail", func(t *testing.T, f bitmask.Flag64, combinationName string) {

		sendMail(t, 1)

	})
}

func sendMail(t *testing.T, count int) {
	for i := 0; i < count; i++ {
		t.Logf("Send %d\n", i)
		// con, err := smtp.Dial(DefaultAddress)
		// assert.NoError(t, err)
		//assert.NotNil(t, con)

		//		assert.NoError(t,con.Hello("abcd"))
		assert.NoError(t, con.Mail(RemoteUser1))

		// assert.Regexp(t, regexp.MustCompile(`^250 OK`), con.Rcpt(LocalUser1Email), "adding a recipient")
		assert.NoError(t, con.Rcpt(LocalUser1Email), "adding a recipient")

		t.Logf("Con %v", con)
		wc, err := con.Data()

		assert.NoError(t, err, "get a writer for sending messages")
		assert.NotNil(t, wc)

		bc, err := wc.Write([]byte(message1))
		assert.NoError(t, err, "sending message")
		assert.Equal(t, bc, len(message1), "whether complete data has been sent")

		//		assert.Regexp(t, regexp.MustCompile(`^250 OK`), wc.Close(), "message confirming receipt of data")
		assert.NoError(t, wc.Close(), "message confirming receipt of data")

	}

}
func TestSendMail(t *testing.T) {
	//logger.Level(logan.TraceLevel)
	runAllCombinations(t, func(t *testing.T, f bitmask.Flag64, combinationName string) {

		if f.HasFlag(ReceiveFromRemoteUsers) {
			assert.NoError(t, con.Mail(RemoteUser1), combinationName)
		} else {
			return
		}
		if !f.HasFlag(ReceiveFromLocalUsers) {
			return
		}

		//		assert.Regexp(t, regexp.MustCompile(`^250 OK`), con.Rcpt(LocalUser1Email), "adding a recipient")
		assert.NoError(t, con.Rcpt(LocalUser1Email), "adding a recipient")

		t.Logf("Con %v", con)
		wc, err := con.Data()

		assert.NoError(t, err, "get a writer for sending message")
		assert.NotNil(t, wc)

		bc, err := wc.Write([]byte(message1))
		assert.NoError(t, err, "sending message")
		assert.Equal(t, bc, len(message1), "whether complete data has been sent")

		// assert.Regexp(t, regexp.MustCompile(`^250 OK`), wc.Close(), "message confirming receipt of data")
		assert.NoError(t, wc.Close(), "message confirming receipt of data")

		return
	})
}

func runAllCombinations(t *testing.T, f combinationHandler) {
	combinations := bitmask.AllCombinations64([]bitmask.FlagInfo64{
		{"Authorization", Authorization},
		{"StartTls", StartTls},
		{"ReceiveFromLocalUsers", ReceiveFromLocalUsers},
		{"ReceiveFromRemoteUsers", ReceiveFromRemoteUsers},
	})
	for _, subset := range combinations {
		name := ""
		var flag bitmask.Flag64
		for _, s := range subset {
			if name != "" {
				name += ","
			}
			name += s.Name
			flag.AddFlag(s.Value)
		}
		cn := fmt.Sprintf("Combination: %v\n", name)
		if !t.Run(cn, func(t *testing.T) {
			combination(t, flag, cn, f)
		}) {

			lastCon.Close()
			return
		}

		lastCon.Close()
		// ctx.Done()
	}

}

var lmtps *Listener
var con *smtp.Client

func combination(t *testing.T, f bitmask.Flag64, combinationName string, fn combinationHandler) {

	_, lmtps, con = connectToTestLmtpServer(t, f)

	if lmtps == nil {
		os.Exit(1)
	}
	defer func() {
		t.Log("closing lmtp server")
		lmtps.Stop()
		wg.Wait()
	}()

	assert.NoError(t, con.Hello("abcd"))

	fn(t, f, combinationName)

	con.Close()
}

func eqErrFromString(t *testing.T, err error, expected string, info string) {
	// because if expected = "" then testify will take it as an error and then we have to compare nil
	if expected == "" {
		assert.NoError(t, err, info)
	} else {
		assert.EqualError(t,
			err,
			expected,
			info,
		)

	}

}

func e(s string, a ...interface{}) error {
	return fmt.Errorf(s, a...)
}
func s(s string, a ...interface{}) string {
	return fmt.Sprintf(s, a...)
}

func stringToErr(val string) error {
	if val == "" {
		return nil
	}

	return fmt.Errorf(val)
}

func stringIf(value bool, ifTrue string, ifElse string) string {
	if value {
		return ifTrue
	} else {
		return ifElse
	}
}

func errorIf(value bool, ifTrue error, ifElse error) error {
	if value {
		return ifTrue
	} else {
		return ifElse
	}
}

func connectToTestLmtpServer(t *testing.T, flags bitmask.Flag64) (ctx context.Context, lmtpServer *Listener, con *smtp.Client) {

	lmtpServer = NewDefaultListenerExt(TestServerAddress, ServerHostname, flags)
	var cfunc context.CancelFunc
	ctx, cfunc = context.WithCancel(context.Background())

	if err := lmtpServer.Listen(); err != nil {
		lmtpServer = nil
	}

	// wg.Add(1)
	go func() {
		//wg.Done()

		assert.NoError(t, lmtpServer.Run(ctx, NewTestSuperVisor(ServerHostname), cfunc), "Start Listenera")

	}()

	return ctx, lmtpServer, redial(t)
}

// reset connections counter
func redial(t *testing.T) *smtp.Client {
	if lastCon != nil {
		if e := lastCon.Close(); e != nil {
			// it's often already closed...
			// t.Logf("error closing previous connection %v", e)
		}
		lastCon = nil

	}
	con, err := smtp.Dial(TestServerAddress)
	assert.NoError(t, err)
	if con == nil {
		os.Exit(1)
	}
	lastCon = con
	return con
}

var lastCon *smtp.Client
var wg sync.WaitGroup
var lmtpServer *Listener
var lmtpSession *SessionImpl

func NewTestSuperVisor(hostname string) *TestSupervisor {
	return &TestSupervisor{
		hostname: hostname,
	}
}

type TestSupervisor struct {
	hostname string
}

func (self *TestSupervisor) GetAbsoluteMaildirPath(mb *api.Mailbox) string {
	//TODO implement me
	panic("implement me")
}

func (self *TestSupervisor) Authorization(username string, password string, service api.Service, useSsl api.Maybe) *api.Mailbox {
	//TODO implement me
	panic("implement me")
}

func (self *TestSupervisor) FindMailbox(name string) *api.Mailbox {
	//TODO implement me
	panic("implement me")
}

func (self *TestSupervisor) GenerateNextSpooledMessageName() string {
	//TODO implement me
	panic("implement me")
}

func (self *TestSupervisor) IsLocalEmail(emailAsciiLowerCase string) api.Maybe {
	//TODO implement me
	for _, v := range localUsers {
		if v.String() == emailAsciiLowerCase {
			return api.Yes
		}
	}
	return api.No
}

func (self *TestSupervisor) MailerDaemonEmailAddress() string {
	return MailerDaemonAddressEmail
}
func (self *TestSupervisor) Hostname() string {
	return self.hostname
}

func (self *TestSupervisor) IsLocalDomain(asciiDomainNameLowerCase string) api.Maybe {
	for _, e := range localUsers {
		if e.Host == asciiDomainNameLowerCase {
			return api.Yes
		}
	}
	return api.No
}

func (self *TestSupervisor) IsLocalAddressEmail(emailAsciiLowerCase AddressEmail) api.Maybe {
	for _, e := range localUsers {
		if e.IsEqual(emailAsciiLowerCase) {
			return api.Yes
		}
	}
	return api.No
}

func (self *TestSupervisor) OpenSession(ci ConnectionInfo, logger *slog.Logger) (session Session, err error) {
	lmtpSession = NewLmtpSession(self, logger)
	return lmtpSession, nil
}

func (self *TestSupervisor) CloseSession(session Session) {
	session.Close()
}

func NewLmtpSession(supervisor api.Supervisor, logger *slog.Logger) *SessionImpl {
	return &SessionImpl{
		supervisor:                  supervisor,
		logger:                      logger,
		username:                    LocalUser1Email,
		password:                    TestPassword,
		allowSendFrom:               []string{LocalUser1Email, LocalUser2Email},
		returnAuthErr:               errors.New("Not implemented Auth"),
		returnNewMessageFromErr:     errors.New("Not implemented New message from"),
		returnAddRecipientErr:       errors.New("Not implemented Add Recipient"),
		returnOnReceivingMessageErr: errors.New("OnReceivingMessageErr"),

		loggedIn: false,
	}
}

type SessionImpl struct {
	supervisor                  api.Supervisor
	logger                      *slog.Logger
	maxMessageSize              uint64
	username                    string
	password                    string
	allowSendFrom               []string
	recipients                  []AddressEmail
	from                        AddressEmail
	returnAuthErr               error
	returnNewMessageFromErr     error
	returnAddRecipientErr       error
	returnOnReceivingMessageErr error

	loggedIn bool
}

func (s SessionImpl) OnAuthorization(username string, password string, service api.Service, useSsl api.Maybe) (bool, error) {
	s.loggedIn = s.username == username && s.password == password

	return s.loggedIn, s.returnAuthErr

}

func (s SessionImpl) MaxMessageSize() uint64 {
	return s.maxMessageSize
}
func (s SessionImpl) IsLoggedIn() bool {
	return s.loggedIn
}

/*
func (s SessionImpl) Authorization(username string, password string, service api.Service, useSsl api.Maybe) (bool, error) {
	s.loggedIn = s.username == username && s.password == password

	return s.loggedIn, s.returnAuthErr
}
*/

// The server knows that I am logged in, the question is whether you can use this address as a sender
func (s SessionImpl) IsAllowSendAs(addressEmailInAsciiLowerCase AddressEmail) (bool, error) {
	for _, e := range s.allowSendFrom {
		if e == addressEmailInAsciiLowerCase.String() {
			return true, nil
		}
	}
	return false, nil
}

func (s SessionImpl) ResetMessageReceivingStatus() error {
	s.from = AddressEmail{}
	s.recipients = nil
	return nil
}

func (s SessionImpl) AcceptMessageFromEmail(senderAscii AddressEmail) error {

	switch senderAscii.String() {
	case SpamerEmail:
		return NewRejectErr(YouSpamerGoAwayRejectErr)
	case ThrotleEmial:
		return NewRejectErr(TooManyConnectionsRejectErr)
	default:
		/*
			for _, v := range localUsers {
				if v.IsEqual(senderAscii) {
					if !s.loggedIn {

					}
				}
			}*/
		return nil

	}

}

func (s SessionImpl) From() AddressEmail {
	return s.from
}

func (s SessionImpl) AcceptRecipient(recipientAscii AddressEmail) error {
	return nil
}

func (s SessionImpl) Close() {
	s.supervisor = nil
	s.logger = nil
	s.loggedIn = false

}

func (s SessionImpl) AcceptMessage(message *mail.Message) error {
	// chec sie nam tu filtrowaC?
	// bez sensu odrazu do skrzynki
	return nil
}

func (s SessionImpl) ProcessDelivery(proxy MessageReceiverProxy, delivery Delivery, hostname string) error {
	if _, _, err := DefaultNewFileFromDelivery(proxy, delivery, hostname, baseStoragePath); err != nil {
		return err
	}

	return nil
}

var baseStoragePath string

func init() {
	if p, err := os.MkdirTemp("", "zoha-temp-listener*"); err != nil {
		panic(err)
	} else {
		baseStoragePath = p
	}
}
