package command

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os/exec"
	"sync"
)

const (
	ON                = true
	OFF               = true
	MaxCommandProcess = 10
)

func (s *Service) Runner() {

	commandLimit := make(chan struct{}, MaxCommandProcess)
	defer close(commandLimit)
	wg := sync.WaitGroup{}
	ctx := context.Background()
	select {

	case <-s.stopSignal:
		return
	default:
		for {
			if l, err := s.ScriptsCache.GetLen(); l > 0 && err == nil {
				scriptIds, _ := s.ScriptsCache.GetAllKeys()
				wg.Add(len(scriptIds))

				consoleMode := OFF
				if len(scriptIds) == 1 {
					consoleMode = ON
				}

				for _, id := range scriptIds {
					commandLimit <- struct{}{}
					go func(id int64) {
						s.executeCommand(ctx, id, consoleMode, &wg, commandLimit)
					}(id)
				}
				wg.Wait()
			}
		}
	}
}

func (s *Service) executeCommand(ctx context.Context, id int64, consoleMode bool, wg *sync.WaitGroup, ch chan struct{}) {
	defer func() {
		_ = s.ScriptsCache.Delete(id)
		_ = s.ExecCmdCache.Delete(id)
	}()

	scanner, cmd, err := s.commandStart(id)
	if err != nil {
		log.Println(err)
		return
	}

	outputScriptCh := make(chan string, 100)
	writeDoneCh := make(chan struct{})

	defer close(writeDoneCh)

	go s.readCommandOutput(scanner, outputScriptCh)
	go s.writeCommandOutput(ctx, id, consoleMode, outputScriptCh, writeDoneCh)
	if err := scanner.Err(); err != nil {
		log.Printf("error: scanning command_id=%d output: %s", id, err)
	}

	err = cmd.Wait()

	<-writeDoneCh
	if err != nil {
		log.Printf("error: command_id=%d %s", id, err)
	} else {
		log.Printf("command_id=%d executed successfully", id)
	}

}

func (s *Service) commandStart(id int64) (*bufio.Scanner, *exec.Cmd, error) {
	val, _ := s.ScriptsCache.Get(id)
	script := val.(string)
	cmd := exec.Command("/bin/sh", "-c", script)
	_ = s.ExecCmdCache.Set(id, cmd)
	stdout, err := cmd.StdoutPipe()

	if err = cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("error: unsuccessful starting command_id = %d: %s", id, err)
	}

	scanner := bufio.NewScanner(stdout)

	return scanner, cmd, nil
}

func (s *Service) readCommandOutput(scanner *bufio.Scanner, outputScriptCh chan string) {
	defer close(outputScriptCh)

	for scanner.Scan() {
		outputScriptCh <- scanner.Text()
	}
}

func (s *Service) writeCommandOutput(ctx context.Context, id int64, consoleMode bool, outputScriptCh chan string, writeDoneCh chan struct{}) {
	defer func() {
		writeDoneCh <- struct{}{}
	}()

	for consoleScriptLine := range outputScriptCh {
		if consoleMode {
			log.Println(consoleScriptLine)
		}
		if err := s.Repository.CreateCommandOutput(ctx, id, consoleScriptLine); err != nil {
			log.Printf("error writing command_id = %d output to database %s", id, err)
		}
	}
}

func (s *Service) StopRunner() {
	s.stopSignal <- struct{}{}
}
