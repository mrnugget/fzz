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
