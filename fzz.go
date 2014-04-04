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

func drawInitialScreen() (err error) {
	_, err = io.WriteString(os.Stdout, ansiEraseDisplay)
	if err != nil {
		return err
	}
	_, err = io.WriteString(os.Stdout, ansiResetCursor)

	return err
}

func setCursorPos(line int, col int) (err error) {
	str := fmt.Sprintf("\033[%d;%dH", line+1, col+1)
	_, err = io.WriteString(os.Stdout, str)
	return err
}

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
	defer setSttyState(&originalSttyState)

	setSttyState(bytes.NewBufferString("cbreak"))
	setSttyState(bytes.NewBufferString("-echo"))

	var input []byte = make([]byte, 0)
	var b []byte = make([]byte, 1)

	// TODO: Do not print results and input to STDOUT, only to TTY, only print
	// selected result to STDOUT

	// TODO: Run command in the background, draw as long as no new input given
	// if input is given, cancel command in background

	// TODO: Only print as many result lines as we have lines on the screen

	for {
		// Clear screen and set cursor to first row, first col
		err = drawInitialScreen()
		if err != nil {
			log.Fatal(err)
		}

		// Print prompt with already typed input
		prompt := fmt.Sprintf(">> %s", input[:len(input)])
		fmt.Printf("%s", prompt)

		if len(input) > 0 {
			fmt.Printf("\n")

			// Print results
			// TODO: move this in a goroutine, give it a quit-channel, stream the output
			// to the current position (line after prompt)
			arg := fmt.Sprintf("%s", input[:len(input)])
			out, err := exec.Command("ag", arg).Output()
			if err != nil {
				log.Fatal(err)
			}
			outb := bytes.NewBuffer(out)
			outr := bufio.NewReader(outb)
			for i := 0; i < int(winRows)-2; i++ {
				line, err := outr.ReadBytes('\n')
				if err != nil && err != io.EOF {
					log.Fatal(err)
				}
				fmt.Printf("%s", line)
			}

			// Jump back to last typing position
			setCursorPos(0, len(prompt))
		}

		os.Stdin.Read(b)
		switch b[0] {
		case 127:
			// Backspace
			if len(input) > 0 {
				input = input[:len(input)-1]
			}
		case 4, 13:
			// Return or Ctrl-D
			fmt.Println("Result:")
			fmt.Printf("%s%s\n", carriageReturn, string(input[:len(input)]))
			return
		default:
			// TODO: Default is wrong here. Only append printable characters to
			// input

			// TODO: Send a signal through the quit channel to the command in the background,
			// to cancel it
			// Everything else
			input = append(input, b...)
		}
	}
}
