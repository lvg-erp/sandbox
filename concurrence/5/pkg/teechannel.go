package teechan

import (
	"concurrence/5/pkg/waitg"
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
func New(numChans int, ttype int) *TeeChan {

	chans := make([]chan int, numChans)
	//вайтгруппы на каждый канал в отдельности
	wgs := make([]waitg.WaitG, numChans)
	for i := range numChans {
		chans[i] = make(chan int)
	}
	var wg waitg.WaitG
	if ttype == Fast {
		for i := range numChans {
			wgs[i] = &waitg.WaitGNormal{}
		}
		wg = &waitg.WaitGStub{}
	}
	if ttype == Slow {
		for i := range numChans {
			wgs[i] = &waitg.WaitGStub{}
		}
		wg = &waitg.WaitGNormal{}
	}

	return &TeeChan{
		chans:    chans,
		numChans: numChans,
		wgs:      wgs,
		wg:       wg,
	}
}

// теперь делаем эту функцию методом класса
func (t *TeeChan) Execute(in chan int) []chan int {
	go func() {
		defer func() {
			for i := range t.numChans {
				go func() {
					t.wgs[i].Wait()
					close(t.chans[i])
				}()
			}
		}()

		for val := range in {
			for i := range t.numChans {
				t.wgs[i].Add(1)
				t.wg.Add(1)
				go func() {
					defer t.wgs[i].Done()
					defer t.wg.Done()

					t.chans[i] <- val
				}()
			}

			t.wg.Wait()
		}

	}()

	return t.chans

}

// используем контекст
//func (t *TeeChan) Execute(ctx context.Context, in chan int) []chan int {
//	var wg sync.WaitGroup // Единая WaitGroup для всех записей
//	go func() {
//		defer func() {
//			wg.Wait() // Ждём завершения всех записей
//			for i := 0; i < t.numChans; i++ {
//				close(t.chans[i]) // Закрываем все каналы
//			}
//		}()
//
//		for {
//			select {
//			case val, ok := <-in:
//				if !ok {
//					return
//				}
//				for i := 0; i < t.numChans; i++ {
//					wg.Add(1)
//					go func(idx int, v int) {
//						defer wg.Done()
//						t.chans[idx] <- v
//					}(i, val)
//				}
//			case <-ctx.Done():
//				return
//			}
//		}
//	}()
//	return t.chans
//}
