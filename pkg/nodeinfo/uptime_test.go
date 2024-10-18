package nodeinfo

import (
	"fmt"
	"strings"
	"testing"
)

const uptimeStr = `11163672.12 126059598.70
`

func Test_splitUptime(t *testing.T) {

	uv, iv, err := splitUptime(uptimeStr)
	if err != nil {
		t.Errorf(err.Error())
	}
	s := strings.Split(uptimeStr, " ")
	uvs := strings.Split(s[0], ".")
	uvi := strings.Split(s[1], ".")
	if fmt.Sprintf("%d", uv) != strings.TrimSpace(uvs[0]) {
		t.Errorf("Oczekiwałem uptime %d jest %s", uv, uvs[0])
	}
	if fmt.Sprintf("%d", iv) != strings.TrimSpace(uvi[0]) {
		t.Errorf("Oczekiwałem idle %d jest %s", iv, uvi[0])
	}

	t.Logf("Uptime %d, Idle %v", uv, iv)

}
