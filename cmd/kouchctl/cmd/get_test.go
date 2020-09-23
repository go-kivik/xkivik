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
	"testing"

	"gitlab.com/flimzy/testy"
)

func Test_get_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("missing document", cmdTest{
		args: []string{"get"},
		err:  "no document specified to get",
	})
	tests.Add("invalid URL on command line", cmdTest{
		args: []string{"-d", "get", "http://localhost:1/foo/bar/%xxx"},
		err:  `parse "http://localhost:1/foo/bar/%xxx": invalid URL escape "%xx"`,
	})
	tests.Add("full url on command line", cmdTest{
		args: []string{"-d", "get", "http://localhost:1/foo/bar"},
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		cmd := rootCmd()

		testCmd(t, cmd, tt)
	})
}
