package mtp

import (
	"github.com/malumar/strutils"
	"net/mail"
	"net/textproto"
	"path/filepath"
	"strings"
)

func NewDeliveryFromClient(client Connection, from, to AddressEmail) Delivery {
	return NewDelivery(from, to)
}

// CloneDelivery cloning delivery of the same message, but to a different mailbox
func CloneDelivery(src *Delivery, toMailBox Maildir) Delivery {
	msgNo := incrementMessageNumber()
	d := Delivery{
		MessageNo: msgNo,
		UniqueId:  uniqueMessageId(msgNo),

		From:          src.From,
		To:            src.To,
		Mailbox:       toMailBox,
		CustomMailbox: src.CustomMailbox,
		IsRecent:      src.IsRecent,
		appendHeader:  append([]string{}, src.appendHeader...),
		Flags:         src.Flags,
	}
	return d
}

func NewDelivery(from, to AddressEmail) Delivery {
	msgNo := incrementMessageNumber()

	mid := uniqueMessageId(msgNo)

	return Delivery{
		MessageNo: msgNo,
		UniqueId:  mid,

		From:          from,
		To:            to,
		Mailbox:       Inbox,
		CustomMailbox: "",
		IsRecent:      true,
		appendHeader:  []string{},
		Flags:         ImapFlag{},
	}
}

type Delivery struct {
	// another message number from the moment the application was launched
	MessageNo uint32
	// unique id of this message
	UniqueId string
	From     AddressEmail
	To       AddressEmail
	// where to deliver, to which folder on disk
	Mailbox Maildir
	// If mailbox = Custom, here is its physical name.
	// If you want to filter mail independently based on folders,
	// set Mailbox = Custom in CustomMailbox = ".Notifications"
	// like in gmail
	CustomMailbox string
	Flags         ImapFlag

	// if true, it will be placed in the "${MAILBOX}/new" folder,
	// otherwise in "${MAILBOX}/cur"
	// do not confuse it with the //Recent flag, which should not be used in ImapFlags
	// because it is a session flag assigned by the IMAP server at the session opening level ( RFC3501 )
	IsRecent bool

	// When the message was forwarded
	IsPassed bool

	// will be added at the very top of the message, headers from the real email cannot be modified
	appendHeader []string
}

func (d *Delivery) SpoolDir(basePath string) string {
	fp := d.Mailbox.Path(basePath, d.CustomMailbox)
	if d.IsRecent {
		fp = filepath.Join(fp, "new")
	} else {
		fp = filepath.Join(fp, "cur")
	}
	return fp
}

//func (d *Delivery) HeaderToString() string {
//	return MimeHeadersToString(d.appendHeader, d.headerOrders)
//}

func (d Delivery) HeaderToString() string {
	return MimeHeadersSliceToString(d.appendHeader)
}

// GetHeader get the first header on the list, if we have an appendHeader,
// we return them - because they are priority, otherwise we return those from the email
func (d Delivery) GetHeader(msg *mail.Message, name string) string {

	for _, h := range d.appendHeader {
		if pos := strings.Index(h, name+":"); pos > -1 {
			return h[pos+1:]
		}
	}

	return msg.Header.Get(name)
}

func (d *Delivery) HeaderContainsPattern(msg *mail.Message, names []string, patterns ...string) bool {
	for _, headerName := range names {
		val := strings.TrimSpace(d.GetHeader(msg, headerName))
		if len(val) == 0 {
			return false
		}
		for _, p := range patterns {
			if strutils.Match(val, p) {
				return true
			}
		}

	}
	return false
}

func (d *Delivery) MatchHeader(msg *mail.Message, name string, matcher func(value string) bool) bool {
	val := d.GetHeader(msg, name)
	return matcher(val)
}

func DeliveryAddHeaderAs(d *Delivery, name, value string, first bool) {
	if first && len(d.appendHeader) > 0 {
		// because the order is important to us, add in a given header is added at the very top - the most important one
		d.appendHeader = append([]string{name + ":" + value}, d.appendHeader...)
	} else {
		d.appendHeader = append(d.appendHeader, name+":"+value)
	}

}

func DeliveryDelHeader(d *Delivery, name string) {
	for i, h := range d.appendHeader {
		if strings.HasPrefix(h, name+":") {
			d.appendHeader[i] = ""
		}
	}
}

// HeaderValues get the first priority headers on the list and then the rest of the message
func (d *Delivery) HeaderValues(msg *mail.Message, key string) (ret []string) {

	for _, h := range d.appendHeader {
		if pos := strings.Index(h, key+":"); pos > -1 {
			ret = append(ret, h[pos:])
			// it cut off
			// [elivered-To:john@example.tld]
			// ret = append(ret, h[pos+1:])
		}
	}

	if msg != nil {
		values := textproto.MIMEHeader(msg.Header).Values(key)
		if len(values) > 0 {
			ret = append(ret, values...)
		}
	}
	//fmt.Println("key2", key, ret)

	return
}

func DeliverySetHeader(d *Delivery, name string, value string, first bool) {
	DeliveryDelHeader(d, name)
	DeliveryAddHeaderAs(d, name, value, first)
}

func (d *Delivery) Close() {

	for k := range d.Flags {
		delete(d.Flags, k)
	}
	d.Flags = nil
	d.appendHeader = nil
}
