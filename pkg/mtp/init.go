package mtp

import (
	"encoding/binary"
	"github.com/malumar/zoha/pkg/bitmask"
	"github.com/malumar/zoha/pkg/nodeinfo"
	"github.com/martinlindhe/base36"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"
)

func incrementMessageNumber() uint32 {
	return atomic.AddUint32(&lastMessageId, 1)

}

// http://cr.yp.to/proto/maildir.html
// Since when running applications (e.g. 2 instances) we could have the same initial id,
// we should somehow guard against this, hence the first 3 tags
// 1. PID							4 bytes
// 2. MESSAGE NO					4 bytes
// 3. -- compressed into 		    8 bytes  and this will be our MessageId
// 1. STARTAPPLICATION in UNIXNANO	8 bytes
// 2. Random bytes					8 bytes
// MARKER							1 byte
// CZASDOSTARCZENIA 				4 bytes
// MESSAGE NO						4 bytes
// MARKER							1 byte
// HASH
//

// Dovecot use 58 bytes
// 1491941793.M41850P8566V0000000000000015I0000000004F3030E_0
// our use 44 bytes:
// 1LF3N0QAQ4NB7Z9WCIEO4SGF67TA64IESGBC3GI9UZGG
// and finaly: ${TIMESTAMP}.${MESSAGEUID}.${HOSTNAME}
// 1491941793.16V3L2YTT241QK2IW2SGK8MSR43R28UZAF9MAGV3M4N4
// root@yournpode:/var/log# cat /proc/sys/kernel/random/boot_id
// a37485bb-05d9-497f-bf8e-8d69eec7a1d2
func uniqueMessageId(msgNo uint32) string {
	// 4 bytes: message sequence number from the moment the application was launched
	// 4 bytes: pid
	// 4 bytes: time without nanoseconds
	// 16 bytes: application instance tag
	time.Now().Nanosecond()
	buf := make([]byte, 28)
	binary.LittleEndian.PutUint64(buf, bitmask.PackTwoUint32(msgNo, uPid32))
	binary.LittleEndian.PutUint32(buf[8:], uint32(time.Now().Unix()))
	copy(uptimeBuf[:], buf[12:])
	return base36.EncodeBytes(buf)
}

// do not read directly
var lastMessageId uint32

// set once per instance
var instanceMark string
var appStarted = time.Now().UnixNano()
var uPid32 = uint32(os.Getpid())

const uptimeBufSize = 16 // 2 x 8 bytes for uint64

var lastTimeMinute uint32

func init() {

	go func() {

		//http.HandleFunc("/stat", func(w http.ResponseWriter, r *http.Request) {
		//	stats := atomic.LoadUint32(&lastTimeMinute)
		//	fmt.Fprintf(w, "throughput %dminute %ds", stats, stats/60)
		//})
		//logger.Error("http server stopped", "err", http.ListenAndServe(":8089", nil))
		//os.Exit(1)
	}()
	go func() {
		var lastCount uint32
		for {
			current := atomic.LoadUint32(&lastMessageId)
			stats := current - lastCount
			lastCount = current
			atomic.StoreUint32(&lastTimeMinute, stats)

			time.Sleep(1 * time.Minute)
		}
	}()
	os.Getpid()

	buf := [2]byte{}
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0xABCD)

	switch buf {
	case [2]byte{0xCD, 0xAB}:

		logger.Info("The machine works in LittleEndian mode")
		NativeEndian = binary.LittleEndian
	case [2]byte{0xAB, 0xCD}:
		logger.Info("The machine works in BigEndian mode")
		NativeEndian = binary.BigEndian
	default:
		panic("Could not determine native endianness.")
	}

	uv, iv, err := nodeinfo.Uptime()
	if err != nil {
		logger.Warn("I set the UPTIME reading error to zero", "err", err)
	}

	n, err := os.Hostname()
	if err != nil {
		logger.Error("Error getting hostname", "err", err)
		os.Exit(1)
	} else {

		hostname = strings.Split(n, ".")[0]
	}
	NativeEndian.PutUint64(uptimeBuf[:], uint64(uv))
	NativeEndian.PutUint64(uptimeBuf[8:], uint64(iv))
	logger.Info("Starting the MTP server", "pid", uPid32, "nodeTag", base36.EncodeBytes(uptimeBuf[:]))

}

var NativeEndian binary.ByteOrder
var uptimeBuf [uptimeBufSize]byte

//var logger = logan.Prefix("mtp")
var logger = slog.Default().With("package", "mtp")

var hostname string
