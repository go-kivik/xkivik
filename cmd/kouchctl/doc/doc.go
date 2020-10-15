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
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/icza/dyno"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/errors"
)

type Doc struct {
	data     string
	file     string
	yamlData string
	yamlFile string
}

func New() *Doc {
	return &Doc{}
}

func (d *Doc) ConfigFlags(pf *pflag.FlagSet) {
	pf.StringVarP(&d.data, "data", "d", "", "JSON document data.")
	pf.StringVarP(&d.file, "data-file", "D", "", "Read document data from the named file. Use - for stdin.")
	pf.StringVarP(&d.yamlData, "yaml-data", "y", "", "YAML document data.")
	pf.StringVarP(&d.yamlFile, "yaml-data-file", "Y", "", "Read document data from the named YAML file. Use - for stdin.")
}

// jsonReader converts an io.Reader into a json.Marshaler.
type jsonReader struct{ io.Reader }

var _ json.Marshaler = (*jsonReader)(nil)

// MarshalJSON returns the reader's contents. If the reader is also an io.Closer,
// it is closed.
func (r *jsonReader) MarshalJSON() ([]byte, error) {
	if c, ok := r.Reader.(io.Closer); ok {
		defer c.Close() // nolint:errcheck
	}
	buf, err := ioutil.ReadAll(r)
	return buf, errors.Code(errors.ErrIO, err)
}

// jsonObject turns an arbitrary object into a json.Marshaler.
type jsonObject struct {
	i interface{}
}

var _ json.Marshaler = &jsonObject{}

func (o *jsonObject) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.i)
}

// Data returns a JSON-marshalable object.
func (d *Doc) Data() (json.Marshaler, error) {
	if d.data != "" {
		return json.RawMessage(d.data), nil
	}
	switch d.file {
	case "-":
		return &jsonReader{os.Stdin}, nil
	case "":
	default:
		f, err := os.Open(d.file)
		return &jsonReader{f}, errors.Code(errors.ErrNoInput, err)
	}
	if d.yamlData != "" {
		return yaml2json(ioutil.NopCloser(strings.NewReader(d.yamlData)))
	}
	switch d.yamlFile {
	case "-":
		return yaml2json(os.Stdin)
	case "":
	default:
		f, err := os.Open(d.yamlFile)
		if err != nil {
			return nil, errors.Code(errors.ErrNoInput, err)
		}
		return yaml2json(f)
	}
	return nil, errors.Code(errors.ErrUsage, "no document provided")
}

func yaml2json(r io.ReadCloser) (json.Marshaler, error) {
	defer r.Close() // nolint:errcheck

	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Code(errors.ErrIO, err)
	}

	var doc interface{}
	if err := yaml.Unmarshal(buf, &doc); err != nil {
		return nil, errors.Code(errors.ErrData, err)
	}
	return &jsonObject{dyno.ConvertMapI2MapS(doc)}, nil
}
