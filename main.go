package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
)

const (
	VERSION                   = "1.0.1"
	defaultPlaceholder        = "{{}}"
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
