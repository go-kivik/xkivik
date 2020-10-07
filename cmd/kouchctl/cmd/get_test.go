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

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/errors"
	"gitlab.com/flimzy/testy"
)

func Test_get_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("missing document", cmdTest{
		args:   []string{"get"},
		status: errors.ErrUsage,
	})
	tests.Add("invalid URL on command line", cmdTest{
		args:   []string{"-d", "get", "http://localhost:1/foo/bar/%xxx"},
		status: errors.ErrUsage,
	})
	tests.Add("full url on command line", cmdTest{
		args:   []string{"-d", "get", "http://localhost:1/foo/bar"},
		status: errors.ErrUnavailable,
	})
	tests.Add("path only on command line", cmdTest{
		args:   []string{"-d", "--kouchconfig", "./testdata/localhost.yaml", "get", "/foo/bar"},
		status: errors.ErrUnavailable,
	})
	tests.Add("document only on command line", cmdTest{
		args:   []string{"-d", "--kouchconfig", "./testdata/localhost.yaml", "get", "bar"},
		status: errors.ErrUnavailable,
	})
	tests.Add("not found", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusNotFound,
		})

		return cmdTest{
			args:   []string{"get", s.URL},
			status: errors.ErrNotFound,
		}
	})
	tests.Add("invalid JSON response", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
				"ETag":         []string{"1-xxx"},
			},
			Body: ioutil.NopCloser(strings.NewReader("invalid")),
		})

		return cmdTest{
			args:   []string{"get", s.URL},
			status: errors.ErrProtocol,
		}
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t)
	})
}
