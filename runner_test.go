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

type TestPipe struct {
	*bytes.Buffer
}

func (t *TestPipe) Close() (err error) {
	return
}

var testInput []byte = []byte{'f', 'o', 'o'}

func TestPrinterIntegration(t *testing.T) {
	buf := new(bytes.Buffer)
	testPrinter := &TestPrinter{
		buffer: buf,
		reset:  false,
	}

	runner := &Runner{
		printer:     testPrinter,
		template:    []string{"echo", "{{}}", "bar"},
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

func TestStreamOutput(t *testing.T) {
	pipe := &TestPipe{bytes.NewBufferString("foo\nbar\n")}
	runner := &Runner{}

	output := make([]string, 0)
	ch := runner.streamOutput(pipe)
	for line := range ch {
		output = append(output, line)
	}

	if len(output) != 2 {
		t.Errorf("streamed output length wrong. expected: %d, actual: %d", 2, len(output))
	}

	if output[0] != "foo\n" {
		t.Errorf("streamed output wrong. expected: %q, actual: %q", "foo\n", output[0])
	}

	if output[1] != "bar\n" {
		t.Errorf("streamed output wrong. expected: %q, actual: %q", "bar\n", output[0])
	}
}

func TestStreamOutputWithoutTrailingNewline(t *testing.T) {
	pipe := &TestPipe{bytes.NewBufferString("foo\nbar")}
	runner := &Runner{}

	output := make([]string, 0)
	ch := runner.streamOutput(pipe)
	for line := range ch {
		output = append(output, line)
	}

	if len(output) != 2 {
		t.Errorf("streamed output length wrong. expected: %d, actual: %d", 2, len(output))
	}

	if output[0] != "foo\n" {
		t.Errorf("streamed output wrong. expected: %q, actual: %q", "foo\n", output[0])
	}

	if output[1] != "bar" {
		t.Errorf("streamed output wrong. expected: %q, actual: %q", "bar", output[0])
	}
}


func TestErrorPrinting(t *testing.T) {
	buf := new(bytes.Buffer)
	testPrinter := &TestPrinter{
		buffer: buf,
		reset:  false,
	}

	runner := &Runner{
		printer:     testPrinter,
		template:    []string{"cat", "{{}}"},
		placeholder: "{{}}",
	}

	runner.runWithInput([]byte{'d', 'o', 'e', 's', 'n', 'o', 't'})

	output := buf.String()
	if output != "cat: doesnot: No such file or directory\n" {
		t.Errorf("unexpected output: %q", output)
	}
}

func TestOutputBuffering(t *testing.T) {
	buf := new(bytes.Buffer)
	testPrinter := &TestPrinter{
		buffer: buf,
		reset:  false,
	}
	outbuf := new(bytes.Buffer)

	runner := &Runner{
		printer:     testPrinter,
		template:    []string{"echo", "{{}}", "bar"},
		placeholder: "{{}}",
	}

	n, err := runner.writeCmdStdout(outbuf)
	if n != 0 {
		t.Errorf("writeCmdStdout with empty buffer wrote wrong number of bytes: %d", n)
	}
	if err != nil {
		t.Errorf("writeCmdStdout with empty buffer failed. error: %s", err)
	}

	runner.runWithInput(testInput)

	printedOutput := buf.String()
	if printedOutput != "foo bar\n" {
		t.Errorf("unexpected printedOutput: %v", printedOutput)
	}

	_, err = runner.writeCmdStdout(outbuf)
	if err != nil {
		t.Errorf("writeCmdStdout failed. error: %s", err)
	}

	currentOutput := outbuf.String()
	if currentOutput != printedOutput {
		t.Errorf("unexpected output: %v", currentOutput)
	}
}

func TestStdin(t *testing.T) {
	buf := new(bytes.Buffer)
	testPrinter := &TestPrinter{
		buffer: buf,
		reset:  false,
	}

	stdinbuf := new(bytes.Buffer)
	stdinbuf.WriteString("foo\nbar")

	runner := &Runner{
		printer:     testPrinter,
		template:    []string{"grep", "{{}}"},
		placeholder: "{{}}",
		stdinbuf:    stdinbuf,
	}

	runner.runWithInput(testInput)
	// Run it two times to see that the stdin is the same every time
	runner.runWithInput(testInput)

	printedOutput := buf.String()
	if printedOutput != "foo\nfoo\n" {
		t.Errorf("unexpected printedOutput: %q", printedOutput)
	}
}
