package main

import (
	"bufio"
	"io"
	"log"
	"os/exec"
	"strings"
)

type Runner struct {
	printer     PrintResetter
	current     *exec.Cmd
	template    string
	placeholder string
}

func (r *Runner) runWithInput(input []byte) {
	cmd := r.cmdWithInput(string(input))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	ch := r.readCmdStdout(stdout)

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	r.current = cmd

	for str := range ch {
		r.printer.Print(str)
	}
	cmd.Wait()
	r.printer.Reset()
}

func (r *Runner) cmdWithInput(input string) *exec.Cmd {
	line := strings.Replace(r.template, r.placeholder, input, -1)
	splitted := strings.Split(line, " ")

	return exec.Command(splitted[0], splitted[1:len(splitted)]...)
}

func (r *Runner) readCmdStdout(stdout io.ReadCloser) <-chan string {
	ch := make(chan string)
	cmdreader := bufio.NewReader(stdout)

	go func() {
		for {
			line, err := cmdreader.ReadBytes('\n')
			if err != nil || err == io.EOF {
				break
			}
			ch <- string(line)
		}
		close(ch)
	}()

	return ch
}

func (r *Runner) killCurrent() {
	if r.current != nil {
		r.current.Process.Kill()
		r.current.Wait()

		r.current = nil
	}

	r.printer.Reset()
}
