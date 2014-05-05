package main

import "bytes"
import "sync"
import "testing"

type TestPipe struct {
	*bytes.Buffer
}

func (t *TestPipe) Close() (err error) {
	return
}

func TestRun(t *testing.T) {
	stdinbuf := &bytes.Buffer{}
	template := []string{"echo", "{{}}"}
	runner := NewRunner(template, "{{}}", "foobar", stdinbuf)

	ch, err := runner.Run()
	if err != nil {
		t.Errorf("runner.Run() returned an error: %s", err)
	}

	output := make([]string, 0)
	for outputline := range ch {
		output = append(output, outputline)
	}
	runner.Wait()

	if len(output) != 1 {
		t.Errorf("output length wrong. expected: %d, actual: %d", 1, len(output))
	}

	if output[0] != "foobar\n" {
		t.Errorf("output wrong. expected: %q, actual: %q", "foobar\n", output[0])
	}
}

func TestStdinbuffer(t *testing.T) {
	stdinbuf := &bytes.Buffer{}
	stdinbuf.WriteString("foo\nbar\n")

	template := []string{"grep", "{{}}"}
	runner := NewRunner(template, "{{}}", "foo", stdinbuf)

	ch, err := runner.Run()
	if err != nil {
		t.Errorf("runner.Run() returned an error: %s", err)
	}

	output := make([]string, 0)
	for outputline := range ch {
		output = append(output, outputline)
	}
	runner.Wait()

	if len(output) != 1 {
		t.Errorf("output length wrong. expected: %d, actual: %d", 1, len(output))
	}

	if output[0] != "foo\n" {
		t.Errorf("output wrong. expected: %q, actual: %q", "foo\n", output[0])
	}
}
func TestStreamOutput(t *testing.T) {
	pipe := &TestPipe{bytes.NewBufferString("foo\nbar\n")}
	wg := &sync.WaitGroup{}
	runner := &Runner{}

	output := make([]string, 0)
	wg.Add(1)
	ch := runner.streamOutput(pipe, wg)
	for line := range ch {
		output = append(output, line)
	}
	wg.Wait()

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
	wg := &sync.WaitGroup{}
	runner := &Runner{}

	output := make([]string, 0)
	wg.Add(1)
	ch := runner.streamOutput(pipe, wg)

	for line := range ch {
		output = append(output, line)
	}
	wg.Wait()

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
