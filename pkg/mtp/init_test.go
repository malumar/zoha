package mtp

import (
	"bytes"
	"encoding/base32"
	"encoding/binary"
	"math"
	"testing"
)

func Test_generateMessageUid(t *testing.T) {

	b := make([]byte, 8)

	binary.LittleEndian.PutUint64(b, math.MaxInt64)
	buf := bytes.NewBuffer(b)
	v, _ := binary.ReadVarint(buf)
	t.Logf("Readed %v, original: %d", int64(v), int64(math.MaxInt64))

	t.Logf("Mid: %s", base32.HexEncoding.EncodeToString(b))
	t.Logf("Mid: %v", uniqueMessageId(1))
	t.Logf("Mid: %v", uniqueMessageId(2))
	t.Logf("Mid: %v", uniqueMessageId(3))
	t.Logf("Mid: %v", uniqueMessageId(4))
	t.Logf("Mid: %v", uniqueMessageId(400))
	t.Logf("Mid: %v", uniqueMessageId(math.MaxUint32-10000))

}
