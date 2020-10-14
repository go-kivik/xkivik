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

// Package doc provides a wrapper for a JSON document, and related command line
// argument handling.
package doc

import (
	"io"
	"os"
	"strings"

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/errors"
	"github.com/spf13/pflag"
)

type Doc struct {
	data string
	file string
}

func New() *Doc {
	return &Doc{}
}

func (d *Doc) ConfigFlags(pf *pflag.FlagSet) {
	pf.StringVarP(&d.data, "data", "d", "", "Document data. Should be valid JSON or YAML.")
	pf.StringVarP(&d.file, "data-file", "D", "", "Read document data from the named file. Use - for stdin.")
}

func (d *Doc) Data() (io.Reader, error) {
	if d.data != "" {
		return strings.NewReader(d.data), nil
	}
	switch d.file {
	case "-":
		return os.Stdin, nil
	case "":
	default:
		f, err := os.Open(d.file)
		return f, errors.Code(errors.ErrNoInput, err)
	}
	return nil, errors.Code(errors.ErrUsage, "no document provided")
}
