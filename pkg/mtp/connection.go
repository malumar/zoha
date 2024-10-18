package mtp

import (
	"github.com/malumar/zoha/pkg/bitmask"
	"time"
)

type Connection interface {
	GetInfo() ConnectionInfo
	GetHello() string
	GetUsername() string
	GetPassword() string
}

func NewConnectionInfo(id int64, flags bitmask.Flag64) ConnectionInfo {
	return ConnectionInfo{
		StartConnection: time.Time{},
		Flags:           flags,
		InstanceId:      0,
		Id:              id,
		UUId:            "",
		LocalAddr:       "",
		RemoteAddr:      "",
		ReverseHostName: "",
	}
}

type State int

const (
	StatePassive State = iota
	StateData
)

type ConnectionInfo struct {
	StartConnection time.Time
	Flags           bitmask.Flag64
	// instance no
	InstanceId int
	// another customer number since launch
	Id int64
	// it should be unique in the long run
	UUId            string
	LocalAddr       string
	RemoteAddr      string
	ReverseHostName string
}
