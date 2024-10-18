package mtp

import (
	"bytes"
	"fmt"
	"testing"
)

func TestMimeHeadersSliceToString(t *testing.T) {
	var out []byte
	writer := bytes.NewBuffer(out)
	WrapText(writer, "Received: by nodename [xxx.xxx.xxx.xxx] with SMTP id 1WA7R8I758PWRAXKCN0WJUPXKDGWW77YY2G4D412C2YO; Mon, 09 Oct 2023 10:52:33 +0200",
		"", "", "", 2)
	str := writer.String()
	for _, c := range str {
		if c == 0x0d {
			fmt.Printf(" %d", c)
		}
	}
}
