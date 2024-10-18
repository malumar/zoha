package mtp

import (
	"fmt"
	"github.com/malumar/zoha/pkg/mtp/headers"
	"testing"
)

func TestDelivery(t *testing.T) {
	d := newd("gosia@studiont")
	DeliveryAddHeaderAs(&d, headers.XEnvelopeFrom, "hey@example.tld", true)
	DeliveryAddHeaderAs(&d, headers.XStoredAt, "node_hostname", true)
	DeliveryAddHeaderAs(&d, headers.XLoop, "johndoe@example.tld", true)

	fmt.Println(">", d.HeaderToString())

	newDelivery := NewDelivery(SanitizeEmailAddress("john@example.tld"), SanitizeEmailAddress("doe@example.tld"))
	DeliveryAddHeaderAs(&newDelivery, headers.XEnvelopeFrom, "hey@example.tld", true)
	fmt.Println(">", newDelivery.HeaderToString())
}

func newd(email string) Delivery {
	lastMessageId++
	uniqId++
	return Delivery{
		MessageNo: uint32(lastMessageId),
		UniqueId:  fmt.Sprintf("%d_%d", uniqId, lastMessageId),

		From:          SanitizeEmailAddress("alice@example.tld"),
		To:            SanitizeEmailAddress(email),
		Mailbox:       Inbox,
		CustomMailbox: "",
		IsRecent:      true,
		appendHeader:  []string{},
		Flags:         ImapFlag{},
	}
}

var lastMsgNo int
var uniqId int
