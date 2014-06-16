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
	"regexp"
	"strings"
	"unicode"
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

func removeLastCharacter(s []byte) []byte {
	if len(s) > 1 {
		r, rsize := utf8.DecodeLastRune(s)
		if r == utf8.RuneError {
			return s[:len(s)-1]
		} else {
			return s[:len(s)-rsize]
		}
	} else if len(s) == 1 {
		return nil
	}
	return s
}

func readCharacter(r io.Reader) <-chan []byte {
	ch := make(chan []byte)
	rs := bufio.NewScanner(r)

	go func() {
		rs.Split(bufio.ScanRunes)

		for rs.Scan() {
			b := rs.Bytes()
			ch <- b
		}
	}()

	return ch
}

func extractInput(args []string, p string) (input string, r []string) {
	hl := len(p) / 2
	expr := fmt.Sprintf("%s(.*)%s", p[:hl], p[hl:])
	matchPlaceholder := regexp.MustCompile(expr)

	for _, arg := range args {
		matches := matchPlaceholder.FindStringSubmatch(arg)
		if len(matches) > 1 {
			input = matches[1]
			r = append(r, p)
		} else {
			r = append(r, arg)
		}
	}

	return
}

type Fzz struct {
	tty           *TTY
	printer       *Printer
	stdinbuf      *bytes.Buffer
	currentRunner *Runner
	input         []byte
	placeholder   string
	args          []string
}

func (fzz *Fzz) Loop() {
	fzz.reset()

	ttych := readCharacter(fzz.tty)
	for {
		if len(fzz.input) > 0 {
			if err := fzz.startNewRunner(); err != nil {
				log.Fatal(err)
			}
		}

		b := <-ttych
		switch b[0] {
		case keyBackspace, keyDelete:
			fzz.input = removeLastCharacter(fzz.input)
		case keyEndOfTransmission, keyLineFeed, keyCarriageReturn:
			if fzz.currentRunner != nil {
				fzz.currentRunner.Wait()
				fzz.tty.resetScreen()
				if len(fzz.input) > 0 {
					io.Copy(os.Stdout, fzz.currentRunner.stdoutbuf)
				}
			} else {
				fzz.tty.resetScreen()
			}
			return
		case keyEscape:
			fzz.tty.resetScreen()
			return
		case keyEndOfTransmissionBlock:
			fzz.input = removeLastWord(fzz.input)
		default:
			if r, _ := utf8.DecodeRune(b); unicode.IsPrint(r) {
				fzz.input = append(fzz.input, b...)
			} else {
				continue
			}
		}

		fzz.killCurrentRunner()

		fzz.reset()
	}
}

func (fzz *Fzz) startNewRunner() error {
	fzz.currentRunner = NewRunner(fzz.args, fzz.placeholder, string(fzz.input), fzz.stdinbuf)
	ch, err := fzz.currentRunner.Run()
	if err != nil {
		return err
	}

	go fzz.printRunnerOutput(ch, utf8.RuneCount(fzz.input))

	return nil
}

func (fzz *Fzz) killCurrentRunner() {
	if fzz.currentRunner != nil {
		go func(runner *Runner) {
			runner.KillWait()
		}(fzz.currentRunner)
	}
}

func (fzz *Fzz) reset() {
	fzz.tty.resetScreen()
	fzz.tty.printPrompt(fzz.input[:len(fzz.input)])
	fzz.printer.Reset()
}

func (fzz *Fzz) printRunnerOutput(out <-chan string, inputlen int) {
	for line := range out {
		fzz.printer.Print(line)
	}
	fzz.tty.cursorAfterPrompt(inputlen)
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

	input, args := extractInput(flag.Args(), placeholder)

	if !containsPlaceholder(args, placeholder) {
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

	fzz := &Fzz{
		printer:     printer,
		tty:         tty,
		stdinbuf:    &stdinbuf,
		input:       []byte(input),
		placeholder: placeholder,
		args:        args,
	}
	fzz.Loop()
}
