package nodeinfo

import (
	"fmt"
	"strconv"
	"strings"
)

func splitUptime(s string) (uptime int64, idle int64, err error) {
	str := strings.Split(s, " ")
	if len(str) != 2 {
		return 0, 0, fmt.Errorf("expected 2 arguments, found %d", len(str))
	}
	arg1 := strings.Split(str[0], ".")
	arg2 := strings.Split(str[1], ".")
	if len(arg1) != 2 || len(arg2) != 2 {
		return 0, 0, fmt.Errorf("expected 2 arguments, found  %d and %d", len(arg1), len(arg2))
	}

	uv, err := strconv.ParseInt(arg1[0], 10, 0)
	if err != nil {
		return 0, 0, fmt.Errorf("uptime parsing error %v", err)
	}
	iv, err := strconv.ParseInt(arg2[0], 10, 0)
	if err != nil {
		return 0, 0, fmt.Errorf("idle parsing error %v", err)
	}

	return uv, iv, nil
}
