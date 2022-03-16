// nolint forbidigo
package testhelper

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/acarl005/stripansi"

	"github.com/spf13/cast"
)

type EnvProvider interface {
	MustGet(key string) string
}

// ReplaceEnvsString replaces ENVs in given string.
func ReplaceEnvsString(str string, provider EnvProvider) string {
	return regexp.
		MustCompile(`%%[a-zA-Z0-9\-_]+%%`).
		ReplaceAllStringFunc(str, func(s string) string {
			return provider.MustGet(strings.Trim(s, `%`))
		})
}

// stripAnsiWriter strips ANSI characters from
type stripAnsiWriter struct {
	buf    *bytes.Buffer
	writer io.Writer
}

func newStripAnsiWriter(writer io.Writer) *stripAnsiWriter {
	return &stripAnsiWriter{
		buf:    &bytes.Buffer{},
		writer: writer,
	}
}

func (w *stripAnsiWriter) writeBuffer() error {
	if _, err := w.writer.Write([]byte(stripansi.Strip(w.buf.String()))); err != nil {
		return err
	}
	w.buf.Reset()
	return nil
}

func (w *stripAnsiWriter) Write(p []byte) (int, error) {
	// Append to the buffer
	n, err := w.buf.Write(p)

	// We can only remove an ANSI escape seq if the whole expression is present.
	// ... so if buffer contains new line -> flush
	if bytes.Contains(w.buf.Bytes(), []byte("\n")) {
		if err := w.writeBuffer(); err != nil {
			return 0, err
		}
	}

	return n, err
}

func (w *stripAnsiWriter) Close() error {
	if err := w.writeBuffer(); err != nil {
		return err
	}
	return nil
}

type nopCloser struct {
	io.Writer
}

func (n *nopCloser) Close() error {
	return nil
}

func TestIsVerbose() bool {
	value := os.Getenv("TEST_VERBOSE")
	if value == "" {
		value = "false"
	}
	return cast.ToBool(value)
}

func VerboseStdout() io.WriteCloser {
	if TestIsVerbose() {
		return newStripAnsiWriter(os.Stdout)
	}

	return &nopCloser{io.Discard}
}

func VerboseStderr() io.WriteCloser {
	if TestIsVerbose() {
		return newStripAnsiWriter(os.Stderr)
	}

	return &nopCloser{io.Discard}
}
