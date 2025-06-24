package waitg

import "sync"

type WaitGNormal = sync.WaitGroup

type WaitGStub struct{}

func (w *WaitGStub) Add(_ int) {}
func (w *WaitGStub) Done()     {}
func (w *WaitGStub) Wait()     {}
