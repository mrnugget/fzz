package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
)

const (
	carriageReturn     = "\015"
	defaultPlaceholder = "{{}}"
)

var originalSttyState bytes.Buffer

func main() {
	flag.Parse()
	if len(flag.Args()) < 2 {
		fmt.Printf("usage: fzz [command with placeholder]")
		os.Exit(1)
	}

	tty, err := NewTTY()
	if err != nil {
		log.Fatal(err)
	}

	err = tty.getSttyState(&originalSttyState)
	if err != nil {
		log.Fatal(err)
	}
	defer tty.setSttyState(&originalSttyState)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		tty.setSttyState(&originalSttyState)
		os.Exit(1)
	}()

	tty.setSttyState(bytes.NewBufferString("cbreak"))
	tty.setSttyState(bytes.NewBufferString("-echo"))

	cmdTemplate := strings.Join(flag.Args(), " ")
	printer := NewPrinter(tty, tty.cols, tty.rows-3)
	runner := &Runner{
		printer:     printer,
		template:    cmdTemplate,
		placeholder: defaultPlaceholder,
	}

	stdinstat, err := os.Stdin.Stat()
	if err != nil {
		log.Fatal(err)
	}
	// os.Stdin is a pipe
	if stdinstat.Mode()&os.ModeNamedPipe != 0 {
		// TODO: maybe use io.ReadAll here, and use []byte as runner.stdinbuf
		stdinbuf := new(bytes.Buffer)
		io.Copy(stdinbuf, os.Stdin)
		runner.stdinbuf = stdinbuf
	} else {
		runner.stdinbuf = nil
	}

	input := make([]byte, 0)
	b := make([]byte, 1)

	for {
		tty.resetScreen()
		tty.printPrompt(input[:len(input)])

		if len(input) > 0 {
			runner.killCurrent()

			go func() {
				runner.runWithInput(input[:len(input)])
				tty.cursorAfterPrompt(len(input))
			}()
		}

		tty.Read(b)
		switch b[0] {
		case 127:
			// Backspace
			if len(input) > 1 {
				input = input[:len(input)-1]
			} else if len(input) == 1 {
				input = nil
			}
		case 4, 10, 13:
			// Ctrl-D, line feed, carriage return
			tty.resetScreen()
			runner.writeCmdStdout(os.Stdout)
			return
		default:
			// TODO: Default is wrong here. Only append printable characters to
			// input
			input = append(input, b...)
		}
	}
}
