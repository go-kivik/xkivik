// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package log

import (
	"fmt"
	"io"
	"strings"
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

func New(out, err io.Writer) Logger {
	return &logger{
		stdout: out,
		stderr: err,
	}
}

func (l *logger) err(line string) {
	fmt.Fprintln(l.stderr, strings.TrimSpace(line))
}

// func (l *logger) out(line string) {
// 	fmt.Fprintln(l.stdout, strings.TrimSpace(line))
// }

func (l *logger) Debug(args ...interface{}) {
	l.err(fmt.Sprint(args...))
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.err(fmt.Sprintf(format, args...))
}
