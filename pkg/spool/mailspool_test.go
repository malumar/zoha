package spool

import (
	"testing"
)

func TestName(t *testing.T) {
	p := New(10, "/tmp", 0750)
	t.Log(p.GenFilename("xgz"))
}
