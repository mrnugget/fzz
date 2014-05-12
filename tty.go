package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

const (
	defaultPrompt    = ">> "
	ansiEraseDisplay = "\033[2J"
	ansiResetCursor  = "\033[H"
)

type winsize struct {
	rows, cols, xpixel, ypixel uint16
}

func NewTTY() (t *TTY, err error) {
	fh, err := os.OpenFile("/dev/tty", os.O_RDWR, 0666)
	if err != nil {
		return
	}
	t = &TTY{File: fh, prompt: defaultPrompt}
	t.getWinsize()
	err = t.saveOriginalState()
	return
}

type TTY struct {
	*os.File
	prompt        string
	rows, cols    int
	originalState []byte
}

func (t *TTY) saveOriginalState() (err error) {
	cmd := exec.Command("stty", "-g")
	cmd.Stdin = t.File

	t.originalState, err = cmd.Output()
	if err != nil {
		return err
	}

	if t.originalState[len(t.originalState)-1] == '\n' {
		t.originalState = t.originalState[:len(t.originalState)-1]
	}

	return
}

func (t *TTY) setSttyState(argv ...string) (err error) {
	cmd := exec.Command("stty", argv...)
	cmd.Stdin = t.File
	cmd.Stdout = t.File
	return cmd.Run()
}

func (t *TTY) resetState() {
	t.setSttyState(string(t.originalState))
}

func (t *TTY) getWinsize() {
	ws := winsize{}
	syscall.Syscall(syscall.SYS_IOCTL,
		t.Fd(), uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&ws)))
	t.rows = int(ws.rows)
	t.cols = int(ws.cols)
}

// Clears the screen and sets the cursor to first row, first column
func (t *TTY) resetScreen() {
	fmt.Fprint(t.File, ansiEraseDisplay+ansiResetCursor)
}

// Print prompt with `in`
func (t *TTY) printPrompt(in []byte) {
	fmt.Fprintf(t.File, t.prompt+"%s", in)
}

// Positions the cursor after the prompt and `inlen` colums to the right
func (t *TTY) cursorAfterPrompt(inlen int) {
	t.setCursorPos(0, len(t.prompt)+inlen)
}

// Sets the cursor to `line` and `col`
func (t *TTY) setCursorPos(line int, col int) {
	fmt.Fprintf(t.File, "\033[%d;%dH", line+1, col+1)
}
