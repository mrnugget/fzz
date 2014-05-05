package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"unicode/utf8"
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
	keyEscape                 = 27
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

func mainLoop(tty *TTY, printer *Printer, stdinbuf *bytes.Buffer) {
	var currentRunner *Runner

	input := make([]byte, 0)
	ttych := make(chan []byte)
	stdoutbuf := &bytes.Buffer{}

	go func() {
		rs := bufio.NewScanner(tty)
		rs.Split(bufio.ScanRunes)

		for rs.Scan() {
			b := rs.Bytes()
			ttych <- b
		}

		log.Fatal(rs.Err())
	}()

	tty.resetScreen()
	tty.printPrompt(input[:len(input)])


	for {
		b := <-ttych

		switch b[0] {
		case keyBackspace, keyDelete:
			if len(input) > 1 {
				r, rsize := utf8.DecodeLastRune(input)
				if r == utf8.RuneError {
					input = input[:len(input)-1]
				} else {
					input = input[:len(input)-rsize]
				}
			} else if len(input) == 1 {
				input = nil
			}
		case keyEndOfTransmission, keyLineFeed, keyCarriageReturn:
			if currentRunner != nil {
				currentRunner.Wait()
			}
			tty.resetScreen()
			io.Copy(os.Stdout, stdoutbuf)
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


		if currentRunner != nil {
			go func(runner *Runner) {
				runner.KillWait()
			}(currentRunner)
		}

		tty.resetScreen()
		tty.printPrompt(input)

		printer.Reset()
		stdoutbuf.Reset()

		if len(input) > 0 {
			currentRunner = NewRunner(flag.Args(), placeholder, string(input), stdinbuf)
			outch, errch, err := currentRunner.Run()
			if err != nil {
				log.Fatal(err)
			}

			// TODO:
			// 1.) give each runner its own stdoutbuf
			// 2.) only return one channel on Run()
			// 3.) this one channel contains stderr and stdout
			// 4.) Put the loop below into runner.Run()
			//	   - each stdout line gets written to runner.stdoutbuf
			//	   - each stdout line gets sent through the one channel Run() returns
			//	   - each stderr line gets sent through the one channel Run() returns
			// 5.) Instead of the loop here, create a new goroutine that uses `range` over the one returned channel to print it
			//     after printing it sets the cursor after the prompt

			go func(inputlen int) {
				defer tty.cursorAfterPrompt(inputlen)
				for {
					select {
					case stdoutline, ok := <-outch:
						if !ok {
							return
						}
						printer.Print(stdoutline)
						stdoutbuf.WriteString(stdoutline)
					case stderrline, ok := <-errch:
						if !ok {
							return
						}
						printer.Print(stderrline)
					}
				}
			}(utf8.RuneCount(input))
		}
	}
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

	stdinbuf := bytes.Buffer{}
	if isPipe(os.Stdin) {
		io.Copy(&stdinbuf, os.Stdin)
	}

	printer := NewPrinter(tty, tty.cols, tty.rows-1) // prompt is one row
	mainLoop(tty, printer, &stdinbuf)
}
