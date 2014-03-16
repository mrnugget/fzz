package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
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

func main() {
	err := getSttyState(&originalSttyState)
	if err != nil {
		log.Fatal(err)
	}
	defer setSttyState(&originalSttyState)

	setSttyState(bytes.NewBufferString("cbreak"))
	setSttyState(bytes.NewBufferString("-echo"))

	var b []byte = make([]byte, 1)
	for {
		os.Stdin.Read(b)
		fmt.Printf("Read character: %s\n", b)
	}
}
