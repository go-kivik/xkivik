package output

import (
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/spf13/pflag"
)

const (
	defaultFormat = "json"
)

// Format is the output format interface.
type Format interface {
	Output(io.Writer, interface{}) error
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
	format string
}

// New returns an output formatter instance.
func New() *Formatter {
	return &Formatter{}
}

// ConfigFlags sets up the CLI flags based on the configured formatters.
func (f *Formatter) ConfigFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&f.format, "output", "o", defaultFormat, "Output format. One of: "+strings.Join(names(), "|"))
}
