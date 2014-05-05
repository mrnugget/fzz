package main

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"strings"
	"sync"
)

func cmdWithInput(template []string, placeholder, input string) *exec.Cmd {
	argv := make([]string, len(template))
	for i := range argv {
		argv[i] = strings.Replace(template[i], placeholder, input, -1)
	}
	return exec.Command(argv[0], argv[1:]...)
}

type Runner struct {
	cmd      *exec.Cmd
	stdinbuf *bytes.Buffer
	wg       *sync.WaitGroup
}

func NewRunner(template []string, placeholder, input string, stdin *bytes.Buffer) *Runner {
	cmd := cmdWithInput(template, placeholder, input)

	return &Runner{
		cmd:      cmd,
		stdinbuf: stdin,
		wg:       &sync.WaitGroup{},
	}
}

func (r *Runner) Run() (<-chan string, <-chan string, error) {
	stdout, err := r.cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	stderr, err := r.cmd.StderrPipe()
	if err != nil {
		stdout.Close()
		return nil, nil, err
	}

	r.wg.Add(2)
	outch := r.streamOutput(stdout, r.wg)
	errch := r.streamOutput(stderr, r.wg)

	if r.stdinbuf.Len() != 0 {
		r.cmd.Stdin = bytes.NewBuffer(r.stdinbuf.Bytes())
	}

	err = r.cmd.Start()
	if err != nil {
		return nil, nil, err
	}

	return outch, errch, nil
}

func (r *Runner) streamOutput(rc io.ReadCloser, wg *sync.WaitGroup) <-chan string {
	ch := make(chan string)
	reader := bufio.NewReader(rc)

	go func() {
		for {
			line, err := reader.ReadBytes('\n')
			if s := string(line); s != "" {
				ch <- s
			}
			if err != nil {
				break
			}
		}
		rc.Close()
		close(ch)
		wg.Done()
	}()

	return ch
}

func (r *Runner) KillWait() {
	r.cmd.Process.Kill()
	r.Wait()
}

func (r *Runner) Wait() {
	r.wg.Wait()
	r.cmd.Wait()
}
