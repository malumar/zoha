package mtp

import (
	"bytes"
	"github.com/malumar/filekit"
	"os"
	"path/filepath"
)

// mailbox i.e. the directory to which to move mail after receiving it
type Maildir string

// Special folders start with a dollar
const (
	Inbox   Maildir = ""
	Spam            = ".SPAM"
	Sent            = ".Sent"
	Draft           = ".Draft"
	Trash           = ".Trash"
	Archive         = ".Archive"
	Custom          = "$custom"
)

const (
	ImapFlagSeen     = "\\Seen"
	ImapFlagAnswered = "\\Answered"
	ImapFlagFlagged  = "\\Flagged"
	ImapFlagDeleted  = "\\Delete"
	ImapFlagDraft    = "\\Draft"
)

type ImapFlag map[string]bool

func (i ImapFlag) Set(key string) {
	if key[:1] == "\\" {
		key = key[1:]
	}
	i[key] = true
}

func (i ImapFlag) IsSet(key string) bool {
	if _, found := i[key]; found {
		return true
	}
	return false
}

func (i ImapFlag) Delete(key string) {
	delete(i, key)
}

func (i ImapFlag) String() string {
	buf := bytes.Buffer{}
	for k, _ := range i {
		if buf.Len() > 0 {
			buf.WriteString(",")
		}
		buf.WriteString(k)
	}
	return buf.String()
}

func (i ImapFlag) Keys() []string {
	r := make([]string, 0)
	for k, _ := range i {
		r = append(r, k)
	}
	return r
}

// Create a folder ${MAILDIR}/new if it does not exist
// and return the full path where you should write
func (m Maildir) CreateNewMailDirIfNotExists(basePath string, customValue string,
	folderPerm, filePerm os.FileMode) (string, error) {

	pth := m.Path(basePath, customValue)

	pthNew := filepath.Join(pth, "new")
	if filekit.IsFileExistsAndIsDir(pthNew) {
		return pthNew, nil
	}

	if err := os.MkdirAll(pthNew, folderPerm); err != nil {
		return "", err
	}

	//  we log less important errors, but I don't stop
	if err := os.Mkdir(filepath.Join(pth, "courierimapkeywords"), folderPerm); err != nil {
		logger.Error(err.Error())
	}
	if err := os.Mkdir(filepath.Join(pth, "cur"), folderPerm); err != nil {
		logger.Error(err.Error())
	}
	if err := os.Mkdir(filepath.Join(pth, "tmp"), folderPerm); err != nil {
		logger.Error(err.Error())
	}
	if err := os.WriteFile(filepath.Join(pth, "maildirfolder"), []byte(""), filePerm); err != nil {
		logger.Error(err.Error())

	}
	if err := os.WriteFile(filepath.Join(pth, "courierimapacl"), []byte("owner aceilrstwx\nadministrators aceilrstwx"), filePerm); err != nil {
		logger.Error(err.Error())
	}

	return pthNew, nil
}

func (m Maildir) Path(basePath string, customValue string) string {
	var pth string

	if m == Custom {
		pth = filepath.Join(basePath, customValue)
	} else {
		pth = filepath.Join(basePath, string(m))
	}

	return pth
}

// clone the value of headers that we add to messages in a given sessio
//func CloneAppendHeaders(delivery Delivery) textproto.MIMEHeader {
//	fmt.Printf("Oryginalne: %v\n", delivery.appendHeader)
//	ret := textproto.MIMEHeader{}
//	for k, v := range delivery.appendHeader {
//		ret[k] = v
//	}
//	fmt.Printf("Skopiowane: %v\n", ret)
//	return ret
//}

// because adding headers would make us add to the message the headers already received in the previous session
//func NewMessageDeliveryFrom(delivery Delivery, appendHeaders textproto.MIMEHeader, emailLowerAsciiString string) Delivery {
//	//	delivery.SetHeader(headers.DeliveredTo, emailLowerAsciiString)
//	//delivery.To = SanitizeEmailAddress(emailLowerAsciiString)
//	delivery.Maildir = ToMailDir(emailLowerAsciiString)
//
//	delivery.appendHeader = textproto.MIMEHeader{}
//	for k, v := range appendHeaders {
//		delivery.appendHeader[k] = v
//	}
//	return delivery
//}
func ToMailDir(emailLowerAsciiString string) Maildir {
	return Maildir(emailLowerAsciiString)
}
