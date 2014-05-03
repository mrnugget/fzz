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
	runner := &Runner{
		template:    flag.Args(),
		placeholder: placeholder,
	}

	input := make([]byte, 0)
	ttych := make(chan []byte)

	go func() {
		rs := bufio.NewScanner(tty)
		rs.Split(bufio.ScanRunes)

		for rs.Scan() {
			b := rs.Bytes()
			ttych <- b
		}

		tty.resetScreen()
		log.Fatal(rs.Err())
	}()

	f, err := os.Create("trace.log")
	if err != nil {
		log.Fatal(err)
	}

	tty.resetScreen()
	tty.printPrompt(input[:len(input)])

	var outch <-chan string
	var errch <-chan string
	var stdoutbuf bytes.Buffer

	for {
		f.WriteString("select\n")
		select {
		case line, ok := <-outch:
			fmt.Fprintf(f, "outch: %q %t\n", line, ok)
			if !ok {
				outch = nil
				fmt.Fprintf(f, "outch closed\n", line, ok)
				tty.cursorAfterPrompt(utf8.RuneCount(input))
			} else {
				stdoutbuf.WriteString(line)
				printer.Print(line)
			}
		case line, ok := <-errch:
			fmt.Fprintf(f, "errch: %q %t\n", line, ok)
			if !ok {
				errch = nil
				fmt.Fprintf(f, "errch closed\n", line, ok)
			} else {
				printer.Print(line)
			}
		case b := <-ttych:
			fmt.Fprintf(f, "ttych: %x\n", b)
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
				if errch != nil && outch != nil {
					for line := range outch {
						printer.Print(line)
						stdoutbuf.WriteString(line)
					}
					for line := range errch {
						stdoutbuf.WriteString(line)
					}
				}
				runner.current.Wait()
				tty.resetScreen()
				io.Copy(os.Stdout, &stdoutbuf)
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

			fmt.Fprintf(f, "ttych: got input\n")
			if runner.current != nil {
				fmt.Fprintf(f, "ttych: close stream\n")
				close(runner.stopstream)

				runner.current.Process.Kill()
				fmt.Fprintf(f, "ttych: kill\n")
				runner.current.Wait()
				fmt.Fprintf(f, "ttych: wait\n")

				runner.current = nil
			}
			fmt.Fprintf(f, "ttych: after reset\n")

			tty.resetScreen()
			tty.printPrompt(input)
			tty.cursorAfterPrompt(utf8.RuneCount(input))
			fmt.Fprintf(f, "ttych: cursor after prompt\n")
			printer.Reset()
			if len(input) > 0 {
				fmt.Fprintf(f, "ttych: rerun")
				var err error
				outch, errch, err = runner.runWithInput(input)
				if err != nil {
					printer.Print("error: " + err.Error())
				}
				stdoutbuf.Reset()
			} else {
				outch, errch = nil, nil
				stdoutbuf.Reset()
			}
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
