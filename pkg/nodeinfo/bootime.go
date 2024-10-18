package nodeinfo

import (
	"os/exec"
	"strings"
	"time"
)

func LastBootTime() string {
	out, err := exec.Command("who", "-b").Output()
	if err != nil {
		panic(err)
	}
	t := strings.TrimSpace(string(out))
	t = strings.TrimPrefix(t, "system boot")
	t = strings.TrimSpace(t)
	return t
}

func Timezone() string {
	out, err := exec.Command("date", "+%Z").Output()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(out))
}

func LastSystemBootTime() (time.Time, error) {
	return time.Parse(`2006-01-02 15:04MST`, LastBootTime()+Timezone())
}
