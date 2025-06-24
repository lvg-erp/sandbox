package teechan

import (
	"concurrence/6/pkg/waitg"
	"context"
	"sync"
)

type TeeChan struct {
	chans    []chan int
	numChans int
	wgs      []waitg.WaitG
	wg       waitg.WaitG
}

const ( //реализация медленной и быстрой занрузки
	Fast = iota
	Slow
)

// выносим всю логику в конструктор
func New(numChans int, wgSlow waitg.WaitG, wgFast waitg.WaitG) *TeeChan {

	chans := make([]chan int, numChans)
	//вайтгруппы на каждый канал в отдельности
	wgs := make([]waitg.WaitG, numChans)
	for i := range numChans {
		chans[i] = make(chan int)
	}

	for i := range numChans {
		wgs[i] = wgFast
	}

	return &TeeChan{
		chans:    chans,
		numChans: numChans,
		wgs:      wgs,
		wg:       wgSlow,
	}
}

// теперь делаем эту функцию методом класса
func (t *TeeChan) Execute(ctx context.Context, in chan int) []chan int {
	var wg sync.WaitGroup // Единая WaitGroup для всех записей
	go func() {
		defer func() {
			wg.Wait() // Ждём завершения всех записей
			for i := 0; i < t.numChans; i++ {
				close(t.chans[i]) // Закрываем все каналы
			}
		}()

		for {
			select {
			case val, ok := <-in:
				if !ok {
					return
				}
				for i := 0; i < t.numChans; i++ {
					wg.Add(1)
					go func(idx int, v int) {
						defer wg.Done()
						t.chans[idx] <- v
					}(i, val)
				}
			case <-ctx.Done():
				return
			}

			t.wg.Wait()

		}
	}()
	return t.chans
}
