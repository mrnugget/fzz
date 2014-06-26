package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

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
	expr := fmt.Sprintf("(.*)%s(.*)%s(.*)", p[:hl], p[hl:])
	matchPlaceholder := regexp.MustCompile(expr)

	for _, arg := range args {
		matches := matchPlaceholder.FindStringSubmatch(arg)
		if len(matches) > 3 {
			input = matches[2]
			r = append(r, matches[1]+p+matches[3])
		} else {
			r = append(r, arg)
		}
	}

	return
}
