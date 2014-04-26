package main

import (
	"fmt"
	"io"
	"sync"
)

type PrintResetter interface {
	Print(string) (int, error)
	Reset()
}

func NewPrinter(target io.Writer, maxCol, maxRow int) *Printer {
	return &Printer{
		target:  target,
		maxCol:  maxCol,
		maxRow:  maxRow,
		printed: 0,
		mutex:   &sync.Mutex{},
	}
}

type Printer struct {
	target  io.Writer
	maxCol  int
	maxRow  int
	printed int
	mutex   *sync.Mutex
}

func (p *Printer) Print(line string) (n int, err error) {
	p.mutex.Lock()
	if p.printed == p.maxRow {
		p.mutex.Unlock()
		return 0, nil
	}

	if p.printed == 0 {
		fmt.Fprintf(p.target, "\n")
	}

	// If we're on the last line, cut the newline character off
	if p.printed == p.maxRow-1 && line[len(line)-1] == '\n' {
		n, err = p.printLine(line[:len(line)-1])
	} else {
		n, err = p.printLine(line)
	}

	if err == nil {
		p.printed++
	}

	p.mutex.Unlock()

func (p *Printer) printLine(line string) (n int, err error) {
	if len(line) > p.maxCol {
		n, err = fmt.Fprintf(p.target, "%s\n", line[:p.maxCol])
	} else {
		n, err = fmt.Fprintf(p.target, "%s", line)
	}
	return
}

func (p *Printer) Reset() {
	p.mutex.Lock()
	p.printed = 0
	p.mutex.Unlock()
}
