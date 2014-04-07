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

func (r *Runner) runWithInput(input []byte, kill <-chan bool) {
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

	for {
		select {
		case str, ok := <-ch:
			if !ok {
				cmd.Wait()
				r.printer.Reset()
				return
			}
			r.printer.Print(str)
		case <-kill:
			cmd.Process.Kill()
			cmd.Wait()
			r.printer.Reset()
			return
		}
	}
}

func (r *Runner) cmdWithInput(input string) *exec.Cmd {
	line := strings.Replace(r.template, "{{}}", input, -1)
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

