package gobinlog

import (
    "sync"
    "github.com/ngaut/log"
)

var count int = 0

type WaitGroupWrapper struct {
	sync.WaitGroup
}

func (w *WaitGroupWrapper) Wrap(cb func()) {
    w.Add(1)
    count++
    log.Infof("Group Wrapper begin [%d]",count)
	go func() {
		cb()
        w.Done()
        log.Infof("Group Wrapper end [%d]",count)
        count--
	}()
}
