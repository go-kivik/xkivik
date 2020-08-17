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
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"gitlab.com/flimzy/testy"
)

func Test_root_RunE(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("unknown flag", cmdTest{
		args: []string{"--bogus"},
		err:  "unknown flag: --bogus",
	})
	tests.Add("unknown command", cmdTest{
		args: []string{"bogus"},
		err:  `unknown command "bogus" for "kouchctl"`,
	})
	tests.Add("Debug long", cmdTest{
		args: []string{"--debug"},
		err:  "no context specified",
	})
	tests.Add("Debug short", cmdTest{
		args: []string{"-d"},
		err:  "no context specified",
	})
	tests.Add("context from config file", cmdTest{
		args: []string{"-d", "--kouchconfig", "./testdata/localhost.yaml"},
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		cmd := rootCmd()

		testCmd(t, cmd, tt)
	})
}

type cmdTest struct {
	args  []string
	stdin string
	err   string
}

func testCmd(t *testing.T, cmd *cobra.Command, tt cmdTest) {
	t.Helper()
	cmd.SetArgs(tt.args)
	var err error
	stdout, stderr := testy.RedirIO(strings.NewReader(tt.stdin), func() {
		err = cmd.Execute()
	})
	if d := testy.DiffText(testy.Snapshot(t, "_stdout"), stdout); d != nil {
		t.Errorf("STDOUT: %s", d)
	}
	if d := testy.DiffText(testy.Snapshot(t, "_stderr"), stderr); d != nil {
		t.Errorf("STDERR: %s", d)
	}
	testy.Error(t, tt.err, err)
}
