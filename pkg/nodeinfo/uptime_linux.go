//go:build linux
// +build linux

package nodeinfo

import (
	"os"
)

// Uptime cat /proc/uptime
// 11163672.12 126059598.70
func Uptime() (uptime int64, idle int64, err error) {
	b, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, 0, err
	}
	return splitUptime(string(b))

}
