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

var testInput []byte = []byte{'f', 'o', 'o'}

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
