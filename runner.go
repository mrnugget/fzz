package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type Runner struct {
	current     *exec.Cmd
	template    []string
	placeholder string
	stdinbuf    bytes.Buffer
	stopstream  chan struct{}
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

	r.stopstream = make(chan struct{})
	outch := r.streamOutput(stdout, r.stopstream)
	errch := r.streamOutput(stderr, r.stopstream)

	if r.stdinbuf.Len() != 0 {
		cmd.Stdin = bytes.NewBuffer(r.stdinbuf.Bytes())
	}

	err = cmd.Start()
	if err != nil {
		return nil, nil, err
	}
	r.current = cmd

	return outch, errch, nil
}

func (r *Runner) cmdWithInput(input string) *exec.Cmd {
	argv := make([]string, len(r.template))
	for i := range argv {
		argv[i] = strings.Replace(r.template[i], r.placeholder, input, -1)
	}

	return exec.Command(argv[0], argv[1:]...)
}

func (r *Runner) streamOutput(stdout io.ReadCloser, stop <-chan struct{}) <-chan string {
	ch := make(chan string)
	cmdreader := bufio.NewReader(stdout)

	go func() {
		for {
			select {
			case <-stop:
				close(ch)
				return
			default:
				line, err := cmdreader.ReadBytes('\n')
				if s := string(line); s != "" {
					ch <- s
				}
				if err != nil || err == io.EOF {
					close(ch)
					return
				}
			}
		}
	}()

	return ch
}

func (r *Runner) killCurrent() {
	if r.current != nil {
		fmt.Printf("lol1")
		close(r.stopstream)

		r.current.Process.Kill()
		r.current.Wait()

		r.current = nil
	}
}

func (r *Runner) wait() {
	if r.current != nil {
		r.current.Wait()
		r.current = nil
	}
}
