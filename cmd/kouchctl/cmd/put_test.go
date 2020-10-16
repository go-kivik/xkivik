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
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

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
