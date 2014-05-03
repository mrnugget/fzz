package main

import (
	"bufio"
	"bytes"
	"io"
	"os/exec"
	"strings"
	"sync"
)

type Runner struct {
	c           *exec.Cmd
	template    []string
	placeholder string
	stdinbuf    bytes.Buffer
	wg          *sync.WaitGroup
}

func (r *Runner) runWithInput(input []byte) (<-chan string, <-chan string, error) {
	cmd := r.cmdWithInput(string(input))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdout.Close()
		return nil, nil, err
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	outch := r.streamOutput(stdout, wg)
	wg.Add(1)
	errch := r.streamOutput(stderr, wg)

	if r.stdinbuf.Len() != 0 {
		cmd.Stdin = bytes.NewBuffer(r.stdinbuf.Bytes())
	}

	err = cmd.Start()
	if err != nil {
		return nil, nil, err
	}

	r.c = cmd
	r.wg = wg

	return outch, errch, nil
}

func (r *Runner) cmdWithInput(input string) *exec.Cmd {
	argv := make([]string, len(r.template))
	for i := range argv {
		argv[i] = strings.Replace(r.template[i], r.placeholder, input, -1)
	}

	return exec.Command(argv[0], argv[1:]...)
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
	if r.c != nil {
		r.c.Process.Kill()
		r.Wait()
	}
}

func (r *Runner) Wait() {
	if r.c != nil {
		r.wg.Wait()
		r.c.Wait()
	}
}
