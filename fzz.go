package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

const (
	ansiEraseDisplay = "\033[2J"
	ansiResetCursor  = "\033[H"
	carriageReturn   = "\015"
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

func getSttyState(state *bytes.Buffer) (err error) {
	cmd := exec.Command("stty", "-g")
	cmd.Stdin = os.Stdin
	cmd.Stdout = state
	return cmd.Run()
}

func setSttyState(state *bytes.Buffer) (err error) {
	cmd := exec.Command("stty", state.String())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func NewTTY() (t *TTY, err error) {
	fh, err := os.OpenFile("/dev/tty", os.O_RDWR, 0666)
	if err != nil {
		return
	}
	t = &TTY{fh, ">> "}
	return
}

type TTY struct {
	*os.File
	prompt string
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

func init() {
	ws := getWinsize()
	winRows = ws.rows
	winCols = ws.cols
}

func main() {
	err := getSttyState(&originalSttyState)
	if err != nil {
		log.Fatal(err)
	}
	// TODO: this needs to be run when the process is interrupted
	defer setSttyState(&originalSttyState)

	setSttyState(bytes.NewBufferString("cbreak"))
	setSttyState(bytes.NewBufferString("-echo"))

	tty, err := NewTTY()
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Clean this up. This is a mess.
	var input []byte = make([]byte, 0)
	var b []byte = make([]byte, 1)
	var out []byte

	for {
		tty.resetScreen()
		tty.printPrompt(input[:len(input)])

		var quit chan bool = make(chan bool)
		if len(input) > 0 {
			go func() {
				var ch chan string = make(chan string)

				arg := fmt.Sprintf("%s", input[:len(input)])
				cmd := exec.Command("ag", arg)
				cmdstdout, err := cmd.StdoutPipe()
				if err != nil {
					log.Fatal(err)
				}
				cmdreader := bufio.NewReader(cmdstdout)

				err = cmd.Start()
				if err != nil {
					log.Fatal(err)
				}

				go func() {
					for {
						line, err := cmdreader.ReadBytes('\n')
						if err != nil || err == io.EOF {
							break
						}
						ch <- string(line)
					}
					close(ch)
				}()

				fmt.Fprintf(tty, "\n")

				printed := 0
				for {
					select {
					case str, ok := <-ch:
						if !ok {
							tty.cursorAfterPrompt(len(input))
							return
						}

						printed++
						if len(str) > int(winCols) {
							fmt.Fprintf(tty, "%s", str[:int(winCols)])
						} else {
							fmt.Fprintf(tty, "%s", str)
						}

						if printed > int(winRows)-3 {
							tty.cursorAfterPrompt(len(input))
							return
						}
					case <-quit:
						cmd.Process.Kill()
						tty.cursorAfterPrompt(len(input))
						return
					}
				}
			}()
		}

		os.Stdin.Read(b)
		switch b[0] {
		case 127:
			// Backspace
			if len(input) > 0 {
				input = input[:len(input)-1]
			}
		case 4, 10, 13:
			// Ctrl-D, line feed, carriage return
			// TODO: this is probably wrong, since we need to wait for the cmd
			// in the bg to finish
			fmt.Fprint(os.Stdout, string(out))
			return
		default:
			// TODO: Default is wrong here. Only append printable characters to
			// input
			input = append(input, b...)
		}

		// Non-blocking sent to quit channel.
		if len(input) > 0 {
			select {
			case quit <- true:
			default:
			}
		}
	}
}
