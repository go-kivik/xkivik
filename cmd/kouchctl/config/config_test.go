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

package config

import (
	"io/ioutil"
	"os"
	"testing"

	"gitlab.com/flimzy/testy"
	"gopkg.in/yaml.v3"

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/log"
)

func Test_unmarshalContext(t *testing.T) {
	type tt struct {
		input string
		err   string
	}

	tests := testy.NewTable()
	tests.Add("invalid YAML", tt{
		input: "- [",
		err:   "yaml: line 1: did not find expected node content",
	})
	tests.Add("long context", tt{
		input: `
name: long
scheme: https
host: localhost:5984
user: admin
password: abc123
database: foo
`,
	})
	tests.Add("invalid DSN", tt{
		input: `
name: short
dsn: https://admin:%xxx@localhost:5984/somedb
`,
		err: `parse "https://admin:%xxx@localhost:5984/somedb": invalid URL escape "%xx"`,
	})
	tests.Add("full DSN", tt{
		input: `
name: short
dsn: https://admin:abc123@localhost:5984/somedb
`,
	})
	tests.Run(t, func(t *testing.T, tt tt) {
		cx := &Context{}
		err := yaml.Unmarshal([]byte(tt.input), cx)
		testy.Error(t, tt.err, err)
		if d := testy.DiffInterface(testy.Snapshot(t), cx); d != nil {
			t.Error(d)
		}
	})
}

func TestNew(t *testing.T) {
	type tt struct {
		filename string
		env      map[string]string
		err      string
	}

	tests := testy.NewTable()
	tests.Add("no config file", tt{})
	tests.Add("permission deined", func(t *testing.T) interface{} {
		f, err := ioutil.TempFile("", "")
		if err != nil {
			t.Fatal(err)
		}
		_ = f.Close()
		t.Cleanup(func() {
			_ = os.RemoveAll(f.Name())
		})
		if err := os.Chmod(f.Name(), 0); err != nil {
			t.Fatal(err)
		}

		return tt{
			filename: f.Name(),
			err:      "open " + f.Name() + ": permission denied",
		}
	})
	tests.Add("file not found", tt{
		filename: "not found",
	})
	tests.Add("env only", tt{
		env: map[string]string{
			"KOUCHDSN": "http://foo.com/",
		},
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		testEnv(t, tt.env)
		l := log.NewTest()
		cf, err := New(tt.filename, l)
		testy.Error(t, tt.err, err)
		if d := testy.DiffInterface(testy.Snapshot(t), cf); d != nil {
			t.Error(d)
		}
		l.Check(t)
	})
}

func testEnv(t *testing.T, env map[string]string) {
	t.Helper()
	t.Cleanup(testy.RestoreEnv())
	os.Clearenv()
	if err := testy.SetEnv(env); err != nil {
		t.Fatal(err)
	}
}

func TestConfig_DSN(t *testing.T) {
	type tt struct {
		cf   *Config
		want string
		err  string
	}

	tests := testy.NewTable()
	tests.Add("no current context", tt{
		cf:  &Config{},
		err: "no context specified",
	})
	tests.Add("context not found", tt{
		cf:  &Config{CurrentContext: "xxx"},
		err: `context "xxx" not found`,
	})
	tests.Add("only one context, no default", tt{
		cf: &Config{
			Contexts: map[string]*Context{
				"foo": {
					Scheme:   "http",
					Host:     "localhost:5984",
					User:     "admin",
					Password: "abc123",
					Database: "_users",
				},
			},
		},
		want: "http://admin:abc123@localhost:5984/_users",
	})
	tests.Add("success", tt{
		cf: &Config{
			Contexts: map[string]*Context{
				"foo": {
					Scheme:   "http",
					Host:     "localhost:5984",
					User:     "admin",
					Password: "abc123",
					Database: "_users",
				},
			},
			CurrentContext: "foo",
		},
		want: "http://admin:abc123@localhost:5984/_users",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		got, err := tt.cf.DSN()
		testy.Error(t, tt.err, err)
		if got != tt.want {
			t.Errorf("Unexpected result: %s", got)
		}
	})
}
