package main

import (
	"fmt"
	"io"
	"strings"
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
	defer p.mutex.Unlock()

	if p.printed == p.maxRow {
		return 0, nil
	}

	if p.printed == 0 {
		fmt.Fprintf(p.target, "\n")
	}

	n, err = p.printLine(line)
	if err == nil {
		p.printed++
	}

	return
}

func (p *Printer) printLine(line string) (n int, err error) {
	line = strings.TrimSuffix(line, "\n")

	num := len(line)
	if num > p.maxCol {
		num = p.maxCol
	}

	format := "%s\n"
	if p.printed == p.maxRow-1 { // do not print newline on last line
		format = "%s"
	}

	n, err = fmt.Fprintf(p.target, format, line[:num])
	return
}

func (p *Printer) Reset() {
	p.mutex.Lock()
	p.printed = 0
	p.mutex.Unlock()
}
