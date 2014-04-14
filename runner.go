package main

import (
	"bufio"
	"bytes"
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
	stdoutbuf   *bytes.Buffer
	stdinbuf    *bytes.Buffer
}

func (r *Runner) runWithInput(input []byte) {
	cmd := r.cmdWithInput(string(input))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	outch := r.streamOutput(stdout)
	errch := r.streamOutput(stderr)

	if r.stdinbuf != nil {
		cmd.Stdin = bytes.NewBuffer(r.stdinbuf.Bytes())
	}

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	r.stdoutbuf = new(bytes.Buffer)
	r.current = cmd

	for str := range outch {
		r.printer.Print(str)
		r.stdoutbuf.WriteString(str)
	}

	err = cmd.Wait()
	if err != nil {
		for str := range errch {
			r.printer.Print(str)
		}
		return
	}

	r.printer.Reset()
}

func (r *Runner) cmdWithInput(input string) *exec.Cmd {
	line := strings.Replace(r.template, r.placeholder, input, -1)
	splitted := strings.Split(line, " ")

	return exec.Command(splitted[0], splitted[1:len(splitted)]...)
}

func (r *Runner) streamOutput(stdout io.ReadCloser) <-chan string {
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

func (r *Runner) writeCmdStdout(out io.Writer) (n int64, err error) {
	return io.Copy(out, r.stdoutbuf)
}

func (r *Runner) killCurrent() {
	if r.current != nil {
		r.current.Process.Kill()
		r.current.Wait()

		r.current = nil
	}

	r.printer.Reset()
}
