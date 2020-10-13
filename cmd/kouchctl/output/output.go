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

	"github.com/spf13/pflag"

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/errors"
)

// Formatter manages output formatting.
type Formatter struct {
	mu      sync.Mutex
	formats map[string]Format

	defaultFormat string
	format        string
	output        string
	overwrite     bool
}

// New returns an output formatter instance.
func New() *Formatter {
	return &Formatter{
		formats: map[string]Format{},
	}
}

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

// Register registers an output formatter.
func (f *Formatter) Register(name string, fmt Format) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.defaultFormat == "" {
		f.defaultFormat = name
	}
	if _, ok := f.formats[name]; ok {
		panic(name + " already registered")
	}
	f.formats[name] = fmt
}

func (f *Formatter) options() []string {
	if len(f.formats) == 0 {
		panic("no formatters regiestered")
	}
	fmts := make([]string, 1, len(f.formats))
	def, ok := f.formats[f.defaultFormat]
	if !ok {
		panic("default format not registered")
	}
	fmts[0] = formatOptions(f.defaultFormat, def)
	for name, fmt := range f.formats {
		if name != f.defaultFormat {
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

// ConfigFlags sets up the CLI flags based on the configured formatters.
func (f *Formatter) ConfigFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&f.format, "format", "f", f.defaultFormat, "Output format. One of: "+strings.Join(f.options(), "|"))
	fs.StringVarP(&f.output, "output", "o", "", "Output file/directory.")
	fs.BoolVarP(&f.overwrite, "overwrite", "F", false, "Overwrite output file")
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
		return f.formats[f.defaultFormat], nil
	}
	args := strings.SplitN(f.format, "=", 2)
	name := args[0]
	if format, ok := f.formats[name]; ok {
		if fmtArg, ok := format.(FormatArg); ok {
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

		return format, nil
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
