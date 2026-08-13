package pkg

import "fmt"

func BannerForMetrics() string {
	return `
	╔═══════════════════════════════════════════════════════════════════╗
	║                    WORKER POOL METRICS                           ║
	╠═══════════════════════════════════════════════════════════════════╣
	║ Uptime:          %-56s ║
	║ Tasks Submitted: %-56d ║
	║ Tasks Completed: %-56d ║
	║ Tasks Failed:    %-56d ║
	║ Tasks Panicked:  %-56d ║
	║ Workers Active:  %-56d ║
	║ Workers Restart: %-56d ║
	║ Queue Size:      %-56d ║
	║ Success Rate:    %-55.1f%% ║
	║ Goroutines:      %-56d ║
	╚═══════════════════════════════════════════════════════════════════╝`
}

func PrintBannerWorker() {
	banner := `
╔══════════════════════════════════════════════════════════════════╗
║                                                                  ║
║     ██╗    ██╗ ██████╗ ██████╗ ██╗  ██╗███████╗██████╗         ║
║     ██║    ██║██╔═══██╗██╔══██╗██║ ██╔╝██╔════╝██╔══██╗        ║
║     ██║ █╗ ██║██║   ██║██████╔╝█████╔╝ █████╗  ██████╔╝        ║
║     ██║███╗██║██║   ██║██╔══██╗██╔═██╗ ██╔══╝  ██╔══██╗        ║
║     ╚███╔███╔╝╚██████╔╝██║  ██║██║  ██╗███████╗██║  ██║        ║
║      ╚══╝╚══╝  ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝        ║
║                                                                  ║
║        Worker Pool with Dynamic Scaling & Health Monitoring      ║
║                                                                  ║
╠══════════════════════════════════════════════════════════════════╣
║  Commands:                                                       ║
║    workers <N>  - Set number of workers                         ║
║    min <N>      - Set minimum workers                          ║
║    max <N>      - Set maximum workers                          ║
║    status       - Show current status                          ║
║    exit         - Exit application                             ║
╚══════════════════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
}
