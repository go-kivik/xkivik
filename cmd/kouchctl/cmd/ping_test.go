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
)

func Test_ping_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("missing document", cmdTest{
		args:   []string{"get"},
		status: 1,
	})
	tests.Add("invalid URL on command line", cmdTest{
		args:   []string{"-d", "ping", "http://localhost:1/foo/bar/%xxx"},
		status: 1,
	})
	tests.Add("full url on command line", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			Body: ioutil.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		})

		return cmdTest{
			args: []string{"ping", s.URL},
		}
	})
	tests.Add("server only on command line", func(t *testing.T) interface{} {
		s := testy.ServeResponse(&http.Response{
			Body: ioutil.NopCloser(strings.NewReader(`{"status":"ok"}`)),
		})

		return cmdTest{
			args: []string{"--kouchconfig", "./testdata/localhost.yaml", "ping", s.URL},
		}
	})
	tests.Add("no server provided", cmdTest{
		args:   []string{"ping", "foo/bar"},
		status: 1,
	})
	tests.Add("network error", cmdTest{
		args:   []string{"ping", "http://localhost:9999/"},
		status: 1,
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		cmd := rootCmd()

		testCmd(t, cmd, tt)
	})
}
