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

package cmd

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/icza/dyno"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/errors"
)

type put struct {
	data string
	file string
	yaml bool

	doc *cobra.Command

	*root
}

func putCmd(r *root) *cobra.Command {
	c := &put{
		root: r,
	}
	c.doc = putDocCmd(r, c)
	cmd := &cobra.Command{
		Use:   "put",
		Short: "Put a resource",
		Long:  `Create or update the named resource`,
		RunE:  c.RunE,
	}

	c.configFlags(cmd.PersistentFlags())

	cmd.AddCommand(c.doc)

	return cmd
}

func (c *put) configFlags(pf *pflag.FlagSet) {
	pf.StringVarP(&c.data, "data", "d", "", "JSON document data.")
	pf.StringVarP(&c.file, "data-file", "D", "", "Read document data from the named file. Use - for stdin. Assumed to be JSON, unless the file extension is .yaml or .yml, or the --yaml flag is used.")
	pf.BoolVar(&c.yaml, "yaml", false, "Treat input data as YAML")
}

func (c *put) RunE(cmd *cobra.Command, args []string) error {
	if c.conf.HasDoc() {
		return c.doc.RunE(cmd, args)
	}
	_, err := c.client()
	if err != nil {
		return err
	}

	return errors.New("xxx")
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

func (c *put) jsonData() (json.Marshaler, error) {
	if !c.yaml {
		if c.data != "" {
			return json.RawMessage(c.data), nil
		}
		switch c.file {
		case "-":
			return &jsonReader{os.Stdin}, nil
		case "":
		default:
			if !strings.HasSuffix(c.file, ".yaml") && !strings.HasSuffix(c.file, ".yml") {
				f, err := os.Open(c.file)
				return &jsonReader{f}, errors.Code(errors.ErrNoInput, err)
			}
		}
	}
	if c.data != "" {
		return yaml2json(ioutil.NopCloser(strings.NewReader(c.data)))
	}
	switch c.file {
	case "-":
		return yaml2json(os.Stdin)
	case "":
	default:
		f, err := os.Open(c.file)
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
