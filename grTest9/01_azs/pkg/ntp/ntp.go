package myntp

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func GetNTPServer() (string, error) {
	cmd := exec.Command("timedatectl", "timesync-status")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("timedatectl error: %v", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "Server:") {
			// Ищем содержимое в скобках
			parts := strings.Split(line, "(")
			if len(parts) >= 2 {
				serverPart := parts[1]
				if closingBracket := strings.Index(serverPart, ")"); closingBracket != -1 {
					return serverPart[:closingBracket], nil
				}
			}
		}
	}

	return "", fmt.Errorf("NTP server not found")
}

func SetNTPServer(server string) error {
	cmd := exec.Command("timedatectl", "set-ntp", "true")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("NTP enable error: %v", err)
	}

	configContent := fmt.Sprintf(`[Time]
NTP=%s
FallbackNTP=0.pool.ntp.org 1.pool.ntp.org 2.pool.ntp.org 3.pool.ntp.org
`, server)

	err := os.WriteFile("/etc/systemd/timesyncd.conf", []byte(configContent), 0644)
	if err != nil {
		return fmt.Errorf("write config error: %v", err)
	}

	cmd = exec.Command("systemctl", "restart", "systemd-timesyncd.service")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("service restart error: %v", err)
	}

	ntp, err := GetNTPServer()
	if err != nil {
		return fmt.Errorf("get NTP server error: %v", err)
	}

	if ntp != server {
		return fmt.Errorf("NTP server does not match configured server")
	}

	return nil
}
