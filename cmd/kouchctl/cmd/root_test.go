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
	"context"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/errors"
	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/log"
)

func Test_root_RunE(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("unknown flag", cmdTest{
		args:   []string{"--bogus"},
		status: errors.ErrFailedToInitialize,
	})
	tests.Add("unknown command", cmdTest{
		args:   []string{"bogus"},
		status: errors.ErrFailedToInitialize,
	})
	tests.Add("Debug long", cmdTest{
		args:   []string{"--debug"},
		status: errors.ErrFailedToInitialize,
	})
	tests.Add("Debug short", cmdTest{
		args:   []string{"-d"},
		status: errors.ErrFailedToInitialize,
	})
	tests.Add("context from config file", cmdTest{
		args: []string{"-d", "--kouchconfig", "./testdata/localhost.yaml"},
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		tt.Test(t)
	})
}

type cmdTest struct {
	args   []string
	stdin  string
	status int
}

func (tt *cmdTest) Test(t *testing.T) {
	t.Helper()
	lg := log.New()
	cmd := rootCmd(lg)

	cmd.SetArgs(tt.args)
	var status int
	stdout, stderr := testy.RedirIO(strings.NewReader(tt.stdin), func() {
		status = execute(context.Background(), lg, cmd)
	})
	if d := testy.DiffText(testy.Snapshot(t, "_stdout"), stdout); d != nil {
		t.Errorf("STDOUT: %s", d)
	}
	if d := testy.DiffText(testy.Snapshot(t, "_stderr"), stderr); d != nil {
		t.Errorf("STDERR: %s", d)
	}
	if tt.status != status {
		t.Errorf("Unexpected exit status. Want %d, got %d", tt.status, status)
	}
}
