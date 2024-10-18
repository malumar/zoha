// Spool is used to queue messages to be delivered to users other than the destination email address
// (e.g. send copy to, alias, notifications it), this is a role of MTA such as Postfix
package spool

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/malumar/filekit"
	"github.com/malumar/zoha/pkg/roundrobin"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

type dummyWorker struct {
	Path string
}

// New manager of the directory pool workers
// @workersCount how many workers yow want
// @basePath base path for spool directory
func New(workersCount int, basePath string, fmode os.FileMode) *Spool {
	filekit.MkDirIfNotExists(basePath, fmode)
	ws := make([]*dummyWorker, workersCount)
	for i := 0; i < workersCount; i++ {

		ws[i] = &dummyWorker{Path: filepath.Join(basePath, fmt.Sprintf("%d/", i))}
		filekit.MkDirIfNotExists(ws[i].Path, fmode)
	}
	rr, _ := roundrobin.NewRr(ws...)
	spool := &Spool{
		workersCount: workersCount,
		rr:           rr,
		__basePath:   basePath,
	}

	return spool
}

type Spool struct {
	workersCount int
	rr           *roundrobin.RoundRobin[dummyWorker]
	// variable only to read where to look
	__basePath string
	lastId     atomic.Uint64
	lastWorker atomic.Uint32
}

func (this *Spool) WorkersCount() int {
	return this.workersCount
}

// AbsolutePaths directory paths to all workers
func (this *Spool) AbsolutePaths() []string {
	ret := make([]string, this.workersCount)
	for i := 0; i < this.workersCount; i++ {
		ret[i] = WorkerPath(this.__basePath, i)
	}
	return ret
}

// GenFilename returns the secure folder name
func (this *Spool) GenFilename(hostname string) string {

	id := this.lastId.Add(1)
	dat := make([]byte, 16)
	ts := time.Now().UnixNano()
	binary.BigEndian.PutUint64(dat, uint64(ts))
	binary.BigEndian.PutUint64(dat[8:], id)
	if len(hostname) == 0 {
		hostname = "eml"
	}
	p := this.rr.Next().Path
	if !filekit.IsFileExists(p) {
		os.Mkdir(p, 0600)
	}

	return filepath.Join(p, hex.EncodeToString(dat)+"."+hostname)
}

func WorkerPath(basePath string, workerId int) string {
	return filepath.Join(basePath, fmt.Sprintf("%d/", workerId))
}
