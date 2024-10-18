// Package roundrobin ...
// https://github.com/thegeekyasian/round-robin-go/blob/ee0e2ad436ff/round_robin.go
// author thegeekyasian
package roundrobin

import (
	"errors"
	"sync/atomic"
)

// ErrorNoObjectsProvided is the error that occurs when no objects are provided.
var ErrorNoObjectsProvided = errors.New("no objects provided")

type RoundRobin[O any] struct {
	objects []*O
	next    uint32
}

// New returns RoundRobin implementation with roundrobin.
func NewRr[O any](objects ...*O) (*RoundRobin[O], error) {
	if len(objects) == 0 {
		return nil, ErrorNoObjectsProvided
	}

	return &RoundRobin[O]{
		objects: objects,
	}, nil
}

// Next returns the next object.
func (r *RoundRobin[O]) Next() *O {
	n := atomic.AddUint32(&r.next, 1)

	if int(n) > len(r.objects) {
		atomic.StoreUint32(&r.next, 0)
		n = 1
	}
	return r.objects[(int(n)-1)%len(r.objects)]
}
