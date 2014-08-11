package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"unicode"
	"unicode/utf8"
)

const (
	keyBackspace              = 8
	keyDelete                 = 127
	keyEndOfTransmission      = 4
	keyLineFeed               = 10
	keyCarriageReturn         = 13
	keyEndOfTransmissionBlock = 23
	keyEscape                 = 27
)

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
