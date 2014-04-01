package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

const (
	ansiEraseDisplay = "\033[2J"
	ansiResetCursor  = "\033[H"
	carriageReturn   = "\015"
)

var originalSttyState bytes.Buffer

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
		case 27:
			// Escape
			fmt.Printf("Escape")
		case 127:
			// Backspace
			fmt.Printf("%s", "\b \b")
			input = input[:len(input)-1]
		case 4, 13:
			// Return or Ctrl-D
			fmt.Println("Result:")
			fmt.Printf("%s%s\n", carriageReturn, string(input[:len(input)]))
		default:
			input = append(input, b...)
			fmt.Printf("%c", b[0])
		}
	}
}
