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
	VERSION                   = "0.0.1"
	defaultPlaceholder        = "{{}}"
	keyBackspace              = 8
	keyDelete                 = 127
	keyEndOfTransmission      = 4
	keyLineFeed               = 10
	keyCarriageReturn         = 13
	keyEndOfTransmissionBlock = 23
	keyEscape		  = 27
)

var placeholder string

var usage = `fzz allows you to run a command interactively.

Usage:

	fzz command

The command MUST include the placeholder '{{}}'.

Arguments:

	-v		Print version and exit
`

func printUsage() {
	fmt.Printf(usage)
}

func isPipe(f *os.File) bool {
	s, err := f.Stat()
	if err != nil {
		return false
	}

	return s.Mode()&os.ModeNamedPipe != 0
}

func containsPlaceholder(s []string, ph string) bool {
	for _, v := range s {
		if strings.Contains(v, ph) {
			return true
		}
	}
	return false
}

func validPlaceholder(p string) bool {
	return len(p)%2 == 0
}

func removeLastWord(s []byte) []byte {
	fields := bytes.Fields(s)
	if len(fields) > 0 {
		r := bytes.Join(fields[:len(fields)-1], []byte{' '})
		if len(r) > 1 {
			r = append(r, ' ')
		}
		return r
	}
	return []byte{}
}

func main() {
	flVersion := flag.Bool("v", false, "Print fzz version and quit")
	flag.Usage = printUsage
	flag.Parse()

	if *flVersion {
		fmt.Printf("fzz %s\n", VERSION)
		os.Exit(0)
	}

	if len(flag.Args()) < 2 {
		fmt.Fprintf(os.Stderr, usage)
		os.Exit(1)
	}

	if placeholder = os.Getenv("FZZ_PLACEHOLDER"); placeholder == "" {
		placeholder = defaultPlaceholder
	}

	if !validPlaceholder(placeholder) {
		fmt.Fprintln(os.Stderr, "Placeholder is not valid, needs even number of characters")
		os.Exit(1)
	}

	if !containsPlaceholder(flag.Args(), placeholder) {
		fmt.Fprintln(os.Stderr, "No placeholder in arguments")
		os.Exit(1)
	}

	tty, err := NewTTY()
	if err != nil {
		log.Fatal(err)
	}
	defer tty.resetState()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		tty.resetState()
		os.Exit(1)
	}()
	tty.setSttyState("cbreak", "-echo")

	printer := NewPrinter(tty, tty.cols, tty.rows-1) // prompt is one row
	runner := &Runner{
		printer:     printer,
		template:    flag.Args(),
		placeholder: placeholder,
	}

	if isPipe(os.Stdin) {
		runner.stdinbuf = new(bytes.Buffer)
		io.Copy(runner.stdinbuf, os.Stdin)
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
		case keyBackspace, keyDelete:
			if len(input) > 1 {
				input = input[:len(input)-1]
			} else if len(input) == 1 {
				input = nil
			}
		case keyEndOfTransmission, keyLineFeed, keyCarriageReturn:
			tty.resetScreen()
			runner.writeCmdStdout(os.Stdout)
			return
		case keyEscape:
			tty.resetScreen()
			return
		case keyEndOfTransmissionBlock:
			input = removeLastWord(input)
		default:
			// TODO: Default is wrong here. Only append printable characters to
			// input
			input = append(input, b...)
		}
	}
}
