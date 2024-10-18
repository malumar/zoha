package mtp

import (
	"bytes"
	"fmt"
	"log/slog"
	"time"
)

// DovecotMessageFilenameV2 Dovecot compatible file name generator
// V2 that's because at the end :2,
//
// https://wiki2.dovecot.org/MailboxFormat/Maildir
// About locks, files should always land in new, otherwise there may be problems, e.g. loss of messages when performing imap operations
// Delivering mails to new/ directory doesn't have any problems, so there's no need for LDAs to support any type of locking.
//
// https://cr.yp.to/proto/maildir.html
// When you move a file from new to cur, you have to change its name from uniq to uniq:info.
// Make sure to preserve the uniq string, so that separate messages can't bump into each other.
// info is morally equivalent to the Status field used by mbox readers.
// It'd be useful to have MUAs agree on the meaning of info, so I'm keeping a list of info semantics. Here it is.
//
// info starting with "1,": Experimental semantics.
//
// info starting with "2,": Each character after the comma is an independent flag.
//
// Flag "P" (passed): the user has resent/forwarded/bounced this message to someone else.
// Flag "R" (replied): the user has replied to this message.
// Flag "S" (seen): the user has viewed this message, though perhaps he didn't read all the way through it.
// Flag "T" (trashed): the user has moved this message to the trash; the trash will be emptied by a later user action.
// Flag "D" (draft): the user considers this message a draft; toggled at user discretion.
// Flag "F" (flagged): user-defined flag; toggled at user discretion.
// New flags may be defined later. Flags must be stored in ASCII order: e.g., "2,FRS".
func DovecotMessageFilenameV2(delivery *Delivery, proxy MessageReceiverProxy, hostname string, finalFileSize int) string {

	buf := bytes.Buffer{}

	for _, name := range delivery.Flags.Keys() {
		switch name {
		case ImapFlagSeen:
			// note seen cannot be used for w files new/ -- this is described in the maidir specification
			// \Seen only the imap server client sets this flag
			buf.WriteString("S")
			slog.Warn("using flag SEEN")
			break
		case ImapFlagDraft:
			buf.WriteString("D")
			break
		case ImapFlagDeleted:
			buf.WriteString("T")
		case ImapFlagFlagged:
			buf.WriteString("F")
			break
		case ImapFlagAnswered:
			buf.WriteString("R")
			break
		}
	}

	if delivery.IsPassed {
		buf.WriteString("P")
	}

	return fmt.Sprintf(
		"%d.%s.%s,S=%d,W=%d:2,"+buf.String(),
		time.Now().Unix(),
		delivery.UniqueId,
		hostname,
		finalFileSize,
		finalFileSize,
	)
}
