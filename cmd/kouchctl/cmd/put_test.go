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
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/errors"
)

func Test_put_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("missing document", cmdTest{
		args:   []string{"put"},
		status: errors.ErrUsage,
	})
	tests.Add("full url on command line", cmdTest{
		args:   []string{"--debug", "put", "http://localhost:1/foo/bar", "-d", "{}"},
		status: errors.ErrUnavailable,
	})
	tests.Add("json data string", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: ioutil.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		}, func(t *testing.T, req *http.Request) {
			defer req.Body.Close() // nolint:errcheck
			if d := testy.DiffAsJSON(testy.Snapshot(t), req.Body); d != nil {
				t.Error(d)
			}
		})

		return cmdTest{
			args: []string{"--debug", "put", s.URL + "/foo/bar", "--data", `{"foo":"bar"}`},
		}
	})
	tests.Add("json data stdin", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: ioutil.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		}, func(t *testing.T, req *http.Request) {
			defer req.Body.Close() // nolint:errcheck
			if d := testy.DiffAsJSON(testy.Snapshot(t), req.Body); d != nil {
				t.Error(d)
			}
		})

		return cmdTest{
			args:  []string{"--debug", "put", s.URL + "/foo/bar", "--data-file", "-"},
			stdin: `{"foo":"bar"}`,
		}
	})
	tests.Add("json data file", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: ioutil.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		}, func(t *testing.T, req *http.Request) {
			defer req.Body.Close() // nolint:errcheck
			if d := testy.DiffAsJSON(testy.Snapshot(t), req.Body); d != nil {
				t.Error(d)
			}
		})

		return cmdTest{
			args:  []string{"--debug", "put", s.URL + "/foo/bar", "--data-file", "./testdata/doc.json"},
			stdin: `{"foo":"bar"}`,
		}
	})
	tests.Add("yaml data string", func(t *testing.T) interface{} {
		s := testy.ServeResponseValidator(t, &http.Response{
			Body: ioutil.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		}, func(t *testing.T, req *http.Request) {
			defer req.Body.Close() // nolint:errcheck
			if d := testy.DiffAsJSON(testy.Snapshot(t), req.Body); d != nil {
				t.Error(d)
			}
		})

		return cmdTest{
			args: []string{"--debug", "put", s.URL + "/foo/bar", "--yaml", "--data", `foo: bar`},
		}
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t)
	})
}

func Test_put_jsonData(t *testing.T) {
	type tt struct {
		args   []string
		stdin  string
		status int
		err    string
	}

	tests := testy.NewTable()
	tests.Add("no doc", tt{
		status: errors.ErrUsage,
		err:    "no document provided",
	})
	tests.Add("stdin", tt{
		args:  []string{"--data-file", "-"},
		stdin: `{"foo":"bar"}`,
	})
	tests.Add("string", tt{
		args: []string{"--data", `{"xyz":123}`},
	})
	tests.Add("file", tt{
		args: []string{"--data-file", "./testdata/doc.json"},
	})
	tests.Add("missing file", tt{
		args:   []string{"--data-file", "./testdata/missing.json"},
		status: errors.ErrNoInput,
		err:    "open ./testdata/missing.json: no such file or directory",
	})
	tests.Add("yaml string", tt{
		args: []string{"--yaml", "--data", `foo: bar`},
	})
	tests.Add("yaml stdin", tt{
		args:  []string{"--yaml", "--data-file", `-`},
		stdin: "foo: 1234",
	})
	tests.Add("yaml file", tt{
		args: []string{"--yaml", "--data-file", `./testdata/doc.yaml`},
	})
	tests.Add("yaml file missing", tt{
		args:   []string{"--yaml", "--data-file", `./testdata/missing.yaml`},
		status: errors.ErrNoInput,
		err:    "open ./testdata/missing.yaml: no such file or directory",
	})
	tests.Add("yaml file extension", tt{
		args: []string{"--data-file", `./testdata/doc.yaml`},
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		c := &put{}
		flags := pflag.NewFlagSet("x", pflag.ContinueOnError)
		c.configFlags(flags)

		set := func(flag *pflag.Flag, value string) error {
			return flags.Set(flag.Name, value)
		}

		if err := flags.ParseAll(tt.args, set); err != nil {
			t.Fatal(err)
		}

		var r json.Marshaler
		var err error
		_, _ = testy.RedirIO(strings.NewReader(tt.stdin), func() {
			r, err = c.jsonData()
		})

		if status := errors.InspectErrorCode(err); status != tt.status {
			t.Errorf("Unexpected error status. Want %d, got %d", tt.status, status)
		}
		testy.Error(t, tt.err, err)

		if d := testy.DiffAsJSON(testy.Snapshot(t), r); d != nil {
			t.Error(d)
		}
	})
}
