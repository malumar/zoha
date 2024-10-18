package mtp

import (
	"encoding/base64"
	"fmt"
	"github.com/malumar/zoha/api"
	"github.com/malumar/zoha/pkg/bitmask"
	"github.com/martinlindhe/base36"
	"io"
	"log/slog"
	"net"
	"net/textproto"
	"strings"
	"time"
)

type Client struct {
	// current customer number at the time of connection
	CurrentClientNo int32
	isReady         bool
	properClose     bool
	WhenConnected   time.Time
	//	Listener                *LocalSmtpServer
	Helo                    string
	ListenerReverseHostname string
	ListenerAppName         string
	Flags                   bitmask.Flag64
	DidStartTlsNotAllowed   bool
	UseEhlo                 bool
	DidHeloCmd              bool
	WeClose                 bool
	Username                string
	Password                string
	RejectInfo              []string
	MaxErrorsPerConnection  int
	MaxRecipientsPerMessage int
	MaxMessageSizeInBytes   int64
	// copy of listener.DataStartReceivingTimeoutDuration
	DataStartReceivingTimeoutDuration time.Duration
	// copy of listener.DataStartReceivingTimeoutDuration because we set it after each data reception
	ProtocolTimeoutDuration time.Duration

	Con  net.Conn
	Text *textproto.Conn

	ErrorsCount int
	Session     Session

	state ConnectionState

	Info   ConnectionInfo
	logger *slog.Logger
}

func (c *Client) GetInfo() ConnectionInfo {
	return c.Info
}

func (c *Client) GetSession() Session {
	return c.Session
}

func (c *Client) GetUsername() string {
	return c.Username
}

func (c *Client) GetPassword() string {
	return c.Password
}

func (c *Client) GetHello() string {
	return c.Helo
}

func (c *Client) resetMessageState() {
	/*
		c.DidDataRecived = false
		c.CallFunc = nil
		c.State = mtp.StatePassive

	*/
	c.state.Close()
	c.state = NewConnectionState(c.Flags)
}

// Cmd is a convenience function that sends a command and returns the response
func (c *Client) Cmd(expectCode int, format string, args ...interface{}) (int, string, error) {
	id, err := c.Text.Cmd(format, args...)

	if err != nil {
		return 0, "", err
	}

	c.Text.StartResponse(id)

	defer c.Text.EndResponse(id)

	code, msg, err := c.Text.ReadResponse(expectCode)
	return code, msg, err
}

func (c *Client) UnknownCommand(cmd string, msg string) {
	if len(cmd) > 0 {
		c.WriteCmd(fmt.Sprintf("502 command \"%s\" not implemented", cmd))
	} else {
		c.WriteCmd(fmt.Sprintf("502 command \"%s\" not implemented", msg))

	}
}

func (c *Client) IsLoggedIn() bool {
	return c.Session != nil && c.Session.IsLoggedIn()
}

func (c *Client) DidHelo() bool {
	if !c.DidHeloCmd {
		c.WriteCmd("503  Bad sequence of Commands: need HELO command")
		c.ErrorsCount++
		return false
	}
	return true
}

func (c *Client) DidMailFrom() bool {
	if !c.state.fromCommandPassed {
		//if !c.state.from.IsValid() {
		c.WriteCmd("503  Bad sequence of Commands")
		c.ErrorsCount++
		return false
	}
	return true

}

func (c *Client) DidRecipient() bool {
	if c.state.receiver.RecipientsCount() < 1 {
		c.WriteCmd("503  Bad sequence of Commands: need RCPT TO command")
		c.ErrorsCount++
		return false
	}
	return true

}

func (c *Client) AuthMe(username, password string) bool {
	//	mb, err := c.Listener.AuthUser(username, password) // to
	// FIXME: support for SSL, as LMTP client you don't know whether the client is using an encrypted connection
	if ok, err := c.Session.OnAuthorization(username, password, api.SmtpOutService, api.DontKnow); err != nil {
		if sended := sendedErrorToClientIfReject(c, err); !sended {
			c.logger.Error("AuthMe Unable to complete command", "err", err)
			c.WriteCmd("451 4.7.1 Unable to complete command")
			c.ErrorsCount++
			return false
		}
	} else {
		if ok {
			c.WriteCmd("235 2.7.0 Authentication successful")
			return true
		} else {
			c.WriteCmd("535 5.7.8 Authentication credentials invalid")
			c.ErrorsCount++
			return false
		}
	}

	return false
}

func (c *Client) AuthLoginGetUsername(cmd string) {
	data, err := base64.StdEncoding.DecodeString(cmd)
	if err == nil {
		c.Username = string(data)
	} else {
		c.logger.Error("decode auth cmd error", "err", err)
	}
	c.state.callFunc = c.AuthLoginGetPassword
	c.WriteCmd(fmt.Sprintf("334 %s", base64.StdEncoding.EncodeToString([]byte("Password:"))))
}

func (c *Client) AuthLoginGetPassword(cmd string) {
	data, err := base64.StdEncoding.DecodeString(cmd)
	if err == nil {
		c.Password = string(data)
	}
	c.state.callFunc = nil

	_ = c.AuthMe(c.Username, c.Password)
}

func (c *Client) ReadLine() (string, error) {

	msg, err := c.Text.ReadLine()
	if err != nil && err != io.EOF {
		c.logger.Error("received", "msg", msg, "err", err)
		if c.state.disconnectReason == "" {
			c.state.disconnectReason = "read error"
		}
		c.WeClose = true
	} else {
		c.logger.Debug("received", "msg", msg, "err", err)
	}
	return msg, err

}
func (c *Client) WriteCmd(s string) (err error) {
	_, err = c.Text.Cmd(s)

	if err != nil && err != io.EOF {
		c.logger.Error("send", "msg", s, "err", err)

		c.WeClose = true
		if c.state.disconnectReason == "" {
			c.state.disconnectReason = "write error"
		}
		c.ErrorsCount++
	} else {
		c.logger.Debug("send", "msg", s)

	}

	return err
}

func (c *Client) generateUid() string {
	b := make([]byte, 16)
	NativeEndian.PutUint64(b, uint64(time.Now().UnixNano()))
	NativeEndian.PutUint64(b[8:], uint64(c.Info.Id))
	v := base36.EncodeBytes(b)
	b = nil
	return v
}

// Takes the arguments proceeding a command and files them
// into a map[string]string after uppercasing each key.  Sample arg
// string:
//		" BODY=8BITMIME SIZE=1024 SMTPUTF8"
// The leading space is mandatory.
func parseArgs(args []string) (map[string]string, error) {
	argMap := map[string]string{}
	for _, arg := range args {
		if arg == "" {
			continue
		}
		m := strings.Split(arg, "=")
		switch len(m) {
		case 2:
			argMap[strings.ToUpper(m[0])] = m[1]
		case 1:
			argMap[strings.ToUpper(m[0])] = ""
		default:
			return nil, fmt.Errorf("failed to parse arg string: %q", arg)
		}
	}
	return argMap, nil
}
