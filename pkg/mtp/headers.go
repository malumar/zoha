package mtp

import (
	"bytes"
	"net/textproto"

	"strings"
)

func MimeHeadersToStringWithOrder(h textproto.MIMEHeader, order []string) string {
	var out []byte
	writer := bytes.NewBuffer(out)
	if h != nil {
		for _, orderedHeaderName := range order {
			if val, found := h[orderedHeaderName]; found {
				if len(val) == 0 {
					continue
				}
				for _, v := range val {
					WrapText(writer, orderedHeaderName+": "+v, "", "", "", 200) // +1 znak na tab + 1 znak na \n
				}
			}
		}
	}

	return writer.String()
}

// MimeHeadersToString it doesn't really matter
func MimeHeadersToString(h textproto.MIMEHeader) string {
	var out []byte
	writer := bytes.NewBuffer(out)
	if h != nil {
		for headerName, val := range h {
			if len(val) == 0 {
				continue
			}
			for _, v := range val {
				WrapText(writer, headerName+": "+v, "", "", "", 98) // +1 znak na tab + 1 znak na \n
			}

		}
	}

	return writer.String()
}

// MimeHeadersSliceToString it doesn't really matter
func MimeHeadersSliceToString(headers []string) string {
	return MimeHeadersSliceToStringLength(headers, 200)
}

func MimeHeadersSliceToStringLength(headers []string, maxLength int) string {
	var out []byte
	writer := bytes.NewBuffer(out)

	for _, val := range headers {
		if len(val) == 0 {
			continue
		}

		pos := strings.Index(val, ":")

		// Outlook crashes when it reaches 98
		WrapText(writer, textproto.CanonicalMIMEHeaderKey(val[:pos])+": "+val[pos+1:],
			"", "", "", maxLength) // +1 znak na tab + 1 znak na \n

	}

	return writer.String()
}
