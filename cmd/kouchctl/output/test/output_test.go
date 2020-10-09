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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/pflag"

	"gitlab.com/flimzy/testy"

	// Formats
	_ "github.com/go-kivik/xkivik/v4/cmd/kouchctl/output/gotmpl"
	_ "github.com/go-kivik/xkivik/v4/cmd/kouchctl/output/json"
	_ "github.com/go-kivik/xkivik/v4/cmd/kouchctl/output/raw"

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/output"
)

func TestOutput(t *testing.T) {
	type tt struct {
		args  []string
		obj   string
		err   string
		check func()
	}

	tests := testy.NewTable()
	tests.Add("defaults", tt{
		obj: `{"x":"y"}`,
	})
	tests.Add("output file", func(t *testing.T) interface{} {
		var dir string
		t.Cleanup(testy.TempDir(t, &dir))
		path := filepath.Join(dir, "test.json")

		return tt{
			args: []string{"-o", path},
			obj:  `{"x":"y"}`,
			check: func() {
				buf, err := ioutil.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if d := testy.DiffAsJSON([]byte(`{"x":"y"}`), buf); d != nil {
					t.Error(d)
				}
			},
		}
	})
	tests.Add("overwrite fail", func(t *testing.T) interface{} {
		var dir string
		t.Cleanup(testy.TempDir(t, &dir))
		path := filepath.Join(dir, "test.json")
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = fmt.Fprintf(f, "asdf")
		_ = f.Close()

		return tt{
			args: []string{"-o", path},
			obj:  `{"x":"y"}`,
			err:  "open " + path + ": file exists",
		}
	})
	tests.Add("overwrite success", func(t *testing.T) interface{} {
		var dir string
		t.Cleanup(testy.TempDir(t, &dir))
		path := filepath.Join(dir, "test.json")
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = fmt.Fprintf(f, "asdf")
		_ = f.Close()

		return tt{
			args: []string{"-o", path, "-O"},
			obj:  `{"x":"y"}`,
			check: func() {
				buf, err := ioutil.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if d := testy.DiffAsJSON([]byte(`{"x":"y"}`), buf); d != nil {
					t.Error(d)
				}
			},
		}
	})
	tests.Add("unsupported format", tt{
		args: []string{"-f", "asdfasdf"},
		err:  "unrecognized output format option: asdfasdf",
	})
	tests.Add("raw", tt{
		args: []string{"-f", "raw"},
		obj:  `{ "x": "y" }`,
	})
	tests.Add("too many args", tt{
		args: []string{"-f", "raw=xxx"},
		err:  "format raw takes no arguments",
	})
	tests.Add("missing required arg", tt{
		args: []string{"-f", "go-template"},
		err:  "format go-template requires an argument",
	})
	tests.Add("json indent", tt{
		args: []string{"-f", "json=\t\t"},
		obj:  `{ "x": "y" }`,
	})
	tests.Add("gotmpl, invalid", tt{
		args: []string{"-f", "go-template={{ .x "},
		err:  "template: :1: unclosed action",
	})
	tests.Add("gotmpl", tt{
		args: []string{"-f", "go-template={{ .x }}"},
		obj:  `{ "x": "y" }`,
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		fmt := output.New()
		flags := pflag.NewFlagSet("x", pflag.ContinueOnError)
		fmt.ConfigFlags(flags)

		set := func(flag *pflag.Flag, value string) error {
			return flags.Set(flag.Name, value)
		}

		if err := flags.ParseAll(tt.args, set); err != nil {
			t.Fatal(err)
		}
		var err error
		stdout, stderr := testy.RedirIO(nil, func() {
			err = fmt.Output(strings.NewReader(tt.obj))
		})

		testy.Error(t, tt.err, err)
		if d := testy.DiffText(testy.Snapshot(t, "_stdout"), stdout); d != nil {
			t.Errorf("STDOUT: %s", d)
		}
		if d := testy.DiffText("", stderr); d != nil {
			t.Errorf("STDERR: %s", d)
		}
		if tt.check != nil {
			tt.check()
		}
	})
}
