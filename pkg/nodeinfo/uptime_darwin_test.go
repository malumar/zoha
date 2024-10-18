package nodeinfo

import "testing"

func TestUptime(t *testing.T) {

	u, i, e := Uptime()
	t.Logf("Uptime %v, %v, %v", u, i, e)

}
