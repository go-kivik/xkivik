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

package output

import (
	"io"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/errors"
	"github.com/spf13/pflag"
)

const defaultFormat = "json"

// Format is the output format interface.
type Format interface {
	Output(io.Writer, io.Reader) error
}

// FormatArg is an optional interface. If implemented by a formatter, it
// may receive an argument.
type FormatArg interface {
	Arg(string) error
	Required() bool
}

var (
	mu      sync.Mutex
	formats map[string]Format
)

// Register registers an output formatter.
func Register(name string, fmt Format) {
	mu.Lock()
	defer mu.Unlock()
	if formats == nil {
		formats = make(map[string]Format)
	}
	if _, ok := formats[name]; ok {
		panic(name + " already registered")
	}
	formats[name] = fmt
}

func options() []string {
	if len(formats) == 0 {
		panic("no formatters regiestered")
	}
	fmts := make([]string, 1, len(formats))
	def, ok := formats[defaultFormat]
	if !ok {
		panic("default format not registered")
	}
	fmts[0] = formatOptions(defaultFormat, def)
	for name, fmt := range formats {
		if name != defaultFormat {
			fmts = append(fmts, formatOptions(name, fmt))
		}
	}
	sort.Strings(fmts[1:])
	return fmts
}

func formatOptions(name string, f Format) string {
	if argFmt, ok := f.(FormatArg); ok {
		if argFmt.Required() {
			return name + "=..."
		}
		return name + "[=...]"
	}
	return name
}

// Formatter manages output formatting.
type Formatter struct {
	format    string
	output    string
	overwrite bool
}

// New returns an output formatter instance.
func New() *Formatter {
	return &Formatter{}
}

// ConfigFlags sets up the CLI flags based on the configured formatters.
func (f *Formatter) ConfigFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&f.format, "format", "f", defaultFormat, "Output format. One of: "+strings.Join(options(), "|"))
	fs.StringVarP(&f.output, "output", "o", "", "Output file/directory.")
	fs.BoolVarP(&f.overwrite, "overwrite", "O", false, "Overwrite output file")
}

func (f *Formatter) Output(r io.Reader) error {
	fmt, err := f.formatter()
	if err != nil {
		return err
	}
	out, err := f.writer()
	if err != nil {
		return err
	}
	return fmt.Output(out, r)
}

func (f *Formatter) formatter() (Format, error) {
	if f.format == "" {
		return formats[defaultFormat], nil
	}
	args := strings.SplitN(f.format, "=", 2)
	name := args[0]
	if fmt, ok := formats[name]; ok {
		if fmtArg, ok := fmt.(FormatArg); ok {
			if fmtArg.Required() && len(args) == 1 {
				return nil, errors.Codef(errors.ErrUsage, "format %s requires an argument", name)
			}
			if len(args) > 1 {
				if err := fmtArg.Arg(args[1]); err != nil {
					return nil, errors.Code(errors.ErrUsage, err)
				}
			}
		} else if len(args) > 1 {
			return nil, errors.Codef(errors.ErrUsage, "format %s takes no arguments", name)
		}

		return fmt, nil
	}

	return nil, errors.Codef(errors.ErrUsage, "unrecognized output format option: %s", name)
}

func (f *Formatter) writer() (io.Writer, error) {
	switch f.output {
	case "", "-":
		return os.Stdout, nil
	}
	return f.createFile(f.output)
}

func (f *Formatter) createFile(path string) (*os.File, error) {
	if f.overwrite {
		return os.Create(path)
	}
	return os.OpenFile(path, os.O_EXCL|os.O_CREATE|os.O_WRONLY, 0o666)
}
