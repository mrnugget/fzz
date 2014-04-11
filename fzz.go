package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"unsafe"
)

const (
	ansiEraseDisplay = "\033[2J"
	ansiResetCursor  = "\033[H"
	carriageReturn   = "\015"
	defaultPrompt    = ">> "
)

var originalSttyState bytes.Buffer
var winRows uint16
var winCols uint16

type winsize struct {
	rows, cols, xpixel, ypixel uint16
}

func getWinsize() winsize {
	ws := winsize{}
	syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(0), uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&ws)))
	return ws
}

func NewTTY() (t *TTY, err error) {
	fh, err := os.OpenFile("/dev/tty", os.O_RDWR, 0666)
	if err != nil {
		return
	}
	t = &TTY{fh, defaultPrompt}
	return
}

type TTY struct {
	*os.File
	prompt string
}

func (t *TTY) getSttyState(state *bytes.Buffer) (err error) {
	cmd := exec.Command("stty", "-g")
	cmd.Stdin = t.File
	cmd.Stdout = state
	return cmd.Run()
}

func (t *TTY) setSttyState(state *bytes.Buffer) (err error) {
	cmd := exec.Command("stty", state.String())
	cmd.Stdin = t.File
	cmd.Stdout = t.File
	return cmd.Run()
}

// Clears the screen and sets the cursor to first row, first column
func (t *TTY) resetScreen() {
	// TODO: this is probably wrong since it does not remove the clutter from
	// the tty, but only pushes it to the top where its hidden
	// Instead of using reset screen, we need to go back and redraw the screen.
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

func init() {
	ws := getWinsize()
	winRows = ws.rows
	winCols = ws.cols
}

func main() {
	tty, err := NewTTY()
	if err != nil {
		log.Fatal(err)
	}

	err = tty.getSttyState(&originalSttyState)
	if err != nil {
		log.Fatal(err)
	}
	defer tty.setSttyState(&originalSttyState)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		tty.setSttyState(&originalSttyState)
		os.Exit(1)
	}()

	tty.setSttyState(bytes.NewBufferString("cbreak"))
	tty.setSttyState(bytes.NewBufferString("-echo"))

	cmdTemplate := "ag {{}}"
	placeholder := "{{}}"

	printer := NewPrinter(tty, int(winCols), int(winRows)-3)

	runner := &Runner{
		printer:     printer,
		template:    cmdTemplate,
		placeholder: placeholder,
		buf:         new(bytes.Buffer),
	}

	input := make([]byte, 0)
	b := make([]byte, 1)

	for {
		tty.resetScreen()
		tty.printPrompt(input[:len(input)])

		if len(input) > 0 {
			runner.killCurrent()

			fmt.Fprintf(tty, "\n")

			go func() {
				runner.runWithInput(input[:len(input)])
				tty.cursorAfterPrompt(len(input))
			}()
		}

		tty.Read(b)
		switch b[0] {
		case 127:
			// Backspace
			if len(input) > 1 {
				input = input[:len(input)-1]
			} else if len(input) == 1 {
				input = nil
			}
		case 4, 10, 13:
			// Ctrl-D, line feed, carriage return
			tty.resetScreen()
			runner.writeCmdStdout(os.Stdout)
			return
		default:
			// TODO: Default is wrong here. Only append printable characters to
			// input
			input = append(input, b...)
		}
	}
}
