package main

import "bytes"
import "fmt"
import "io"
import "testing"

type TestPrinter struct {
	buffer io.Writer
	reset  bool
}

func (t *TestPrinter) Print(line string) (n int, err error) {
	return fmt.Fprint(t.buffer, line)
}

func (t *TestPrinter) Reset() {
	t.reset = true
}

func TestPrinterIntegration(t *testing.T) {
	buf := new(bytes.Buffer)
	testPrinter := &TestPrinter{
		buffer: buf,
		reset:  false,
	}

	testInput := []byte{'f', 'o', 'o'}

	runner := &Runner{
		printer:     testPrinter,
		template:    "echo {{}} bar",
		placeholder: "{{}}",
	}

	runner.runWithInput(testInput)

	output := buf.String()
	if output != "foo bar\n" {
		t.Errorf("unexpected output: %v", output)
	}

	if !testPrinter.reset {
		t.Errorf("printer not reset")
	}
}

func TestOutputBuffering(t *testing.T) {
	printbuf := new(bytes.Buffer)
	testPrinter := &TestPrinter{
		buffer: printbuf,
		reset:  false,
	}

	testInput := []byte{'f', 'o', 'o'}

	runner := &Runner{
		printer:     testPrinter,
		template:    "echo {{}} bar",
		placeholder: "{{}}",
	}

	runner.runWithInput(testInput)

	printedOutput := printbuf.String()
	if printedOutput != "foo bar\n" {
		t.Errorf("unexpected printedOutput: %v", printedOutput)
	}

	currentbuf := new(bytes.Buffer)
	n, err := runner.writeCmdStdout(currentbuf)
	if err != nil {
		t.Errorf("writeCmdStdout failed. error: %s", err)
	}

	if n != int64(len(printedOutput)) {
		t.Errorf("writeCmdStdout: wrong number of bytes")
	}

	currentOutput := currentbuf.String()
	if currentOutput != "foo bar\n" {
		t.Errorf("unexpected output: %v", currentOutput)
	}
}
