package log

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

type Logger interface {
	Debug(...interface{})
	Debugf(string, ...interface{})
}

type logger struct {
	stdout io.Writer
	stderr io.Writer
}

var _ Logger = &logger{}

func New(cmd *cobra.Command) Logger {
	return &logger{
		stdout: cmd.OutOrStdout(),
		stderr: cmd.ErrOrStderr(),
	}
}

func (l *logger) err(line string) {
	fmt.Fprintln(l.stderr, strings.TrimSpace(line))
}

func (l *logger) out(line string) {
	fmt.Fprintln(l.stdout, strings.TrimSpace(line))
}

func (l *logger) Debug(args ...interface{}) {
	l.err(fmt.Sprint(args...))
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.err(fmt.Sprintf(format, args...))
}
