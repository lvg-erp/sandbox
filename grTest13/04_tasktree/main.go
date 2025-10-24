package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Task struct {
	ID       int
	SubTasks []Task
}

func main() {

	rootTask := Task{
		ID: 1,
		SubTasks: []Task{
			{ID: 2, SubTasks: []Task{{ID: 4}, {ID: 5}}},
			{ID: 3, SubTasks: []Task{{ID: 6}, {ID: 7, SubTasks: []Task{{ID: 8}}}}},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var (
		completeTask int32
		wg           sync.WaitGroup
		done         = make(chan struct{})
	)

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Printf("Завершено задач: %d (отмена по таймауту)\n", atomic.LoadInt32(&completeTask))
				return
			case <-ticker.C:
				fmt.Printf("Промежуточная статистика выполнено задач: %d\n", atomic.LoadInt32(&completeTask))
			case <-done:
				fmt.Printf("Финальная статистика выполнено задач: %d\n", atomic.LoadInt32(&completeTask))
				return
			}
		}
	}()

	wg.Add(1)
	go processTasks(ctx, rootTask, &completeTask, &wg)

	wg.Wait()
	fmt.Println("Закрываем канал done")
	close(done)
	time.Sleep(100 * time.Millisecond) // Даем время горутине обработать done
}

func processTasks(ctx context.Context, task Task, counter *int32, wg *sync.WaitGroup) {
	defer wg.Done()

	select {
	case <-ctx.Done():
		return
	default:
	}

	time.Sleep(time.Duration(task.ID) * 500 * time.Millisecond)
	atomic.AddInt32(counter, 1)
	fmt.Printf("Завершена задача: %d\n", task.ID)

	for _, task := range task.SubTasks {
		select {
		case <-ctx.Done():
			return
		default:
			wg.Add(1)
			go processTasks(ctx, task, counter, wg)
		}
	}

}
