//go:build darwin || dragonfly || freebsd || netbsd || openbsd
// +build darwin dragonfly freebsd netbsd openbsd

package nodeinfo

import (
	"encoding/binary"
	"golang.org/x/sys/unix"
)

// konwersja time.Unix(wynik, 0)
func Uptime() (uptime int64, idle int64, err error) {
	str, errs := unix.Sysctl("kern.boottime")

	if errs != nil {
		return 0, 0, errs
	}

	data := binary.LittleEndian.Uint32([]byte(str))

	return int64(int32(data)), 0, nil

}
