package main

import (
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
	str := fmt.Sprintf("\033[%d;%dH", line + 1, col + 1)
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

	err = drawInitialScreen()
	if err != nil {
		log.Fatal(err)
	}

	var input []byte = make([]byte, 0)
	var b []byte = make([]byte, 1)
	for {
		os.Stdin.Read(b)
		switch b[0] {
		case 127:
			// Backspace
			fmt.Printf("%s", "\b \b")
			input = input[:len(input)-1]
		case 4, 13:
			// Return or Ctrl-D
			fmt.Println("Result:")
			fmt.Printf("%s%s\n", carriageReturn, string(input[:len(input)]))
		default:
			// Everything else
			input = append(input, b...)
			fmt.Printf("%c", b[0])
		}

		iname := fmt.Sprintf("*%s*", input[:len(input)])
		out, err := exec.Command("find", ".", "-iname", iname).Output()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("\n%s", out)
		setCursorPos(0, len(input))
	}
}
