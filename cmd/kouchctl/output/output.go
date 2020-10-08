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

func names() []string {
	if len(formats) == 0 {
		panic("no formatters regiestered")
	}
	fmts := make([]string, 1, len(formats))
	if _, ok := formats[defaultFormat]; !ok {
		panic("default format not registered")
	}
	fmts[0] = defaultFormat
	for name := range formats {
		if name != defaultFormat {
			fmts = append(fmts, name)
		}
	}
	sort.Strings(fmts[1:])
	return fmts
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
	fs.StringVarP(&f.format, "format", "f", defaultFormat, "Output format. One of: "+strings.Join(names(), "|"))
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
	if fmt, ok := formats[f.format]; ok {
		return fmt, nil
	}

	return nil, errors.Codef(errors.ErrUsage, "unrecognized output format option: %s", f.format)
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
