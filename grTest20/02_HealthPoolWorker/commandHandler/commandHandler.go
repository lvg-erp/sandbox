package commandHandler

import (
	"bufio"
	"context"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"workerpool/metrics"
	"workerpool/workerPool"
)

// ===== КОМАНДЫ ДЛЯ УПРАВЛЕНИЯ =====

type CommandHandler struct {
	pool    *workerPool.WorkerPool
	metrics *metrics.Metrics
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewCommandHandler(pool *workerPool.WorkerPool, metrics *metrics.Metrics) *CommandHandler {
	ctx, cancel := context.WithCancel(context.Background())
	return &CommandHandler{
		pool:    pool,
		metrics: metrics,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (h *CommandHandler) Start() {
	go h.handleCommands()
}

func (h *CommandHandler) handleCommands() {
	scanner := bufio.NewScanner(os.Stdin)

	log.Println("📟 Commands:")
	log.Println("  workers <N>     - Set number of workers")
	log.Println("  min <N>         - Set minimum workers")
	log.Println("  max <N>         - Set maximum workers")
	log.Println("  status          - Show status")
	log.Println("  exit            - Exit application")
	log.Println()

	for {
		select {
		case <-h.ctx.Done():
			return
		default:
			if !scanner.Scan() {
				return
			}

			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			parts := strings.Fields(line)
			if len(parts) == 0 {
				continue
			}

			cmd := parts[0]

			switch cmd {
			case "workers", "w":
				if len(parts) < 2 {
					log.Println("Usage: workers <N>")
					continue
				}
				n, err := strconv.Atoi(parts[1])
				if err != nil {
					log.Printf("Invalid number: %v", err)
					continue
				}
				if err := h.pool.SetWorkers(n); err != nil {
					log.Printf("Error: %v", err)
				} else {
					log.Printf("✅ Workers set to %d", n)
				}

			case "min":
				if len(parts) < 2 {
					log.Println("Usage: min <N>")
					continue
				}
				n, err := strconv.Atoi(parts[1])
				if err != nil {
					log.Printf("Invalid number: %v", err)
					continue
				}
				h.pool.SetMinWorkers(n)
				log.Printf("✅ Min workers set to %d", n)

			case "max":
				if len(parts) < 2 {
					log.Println("Usage: max <N>")
					continue
				}
				n, err := strconv.Atoi(parts[1])
				if err != nil {
					log.Printf("Invalid number: %v", err)
					continue
				}
				h.pool.SetMaxWorkers(n)
				log.Printf("✅ Max workers set to %d", n)

			case "status", "s":
				log.Printf("\n📊 Status:\n")
				log.Printf("   Workers: %d (target: %d, min: %d, max: %d)",
					h.pool.GetWorkerCount(),
					h.pool.GetTargetWorkerCount(),
					h.pool.GetMinWorkers(),
					h.pool.GetMaxWorkers())
				log.Printf("   Queue: %d", h.pool.QueueSize())
				log.Printf("   Goroutines: %d", runtime.NumGoroutine())

			case "exit", "quit", "q":
				log.Println("👋 Exiting...")
				h.cancel()
				return

			default:
				log.Printf("Unknown command: %s", cmd)
				log.Println("Available commands: workers, min, max, status, exit")
			}
		}
	}
}

func (h *CommandHandler) Stop() {
	h.cancel()
}
