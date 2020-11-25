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
	"bytes"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/go-kivik/kivik/v4"

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/output"
	v "github.com/go-kivik/xkivik/v4/cmd/kouchctl/version"
)

type version struct {
	*root
	clientOnly bool
	serverOnly bool
}

func versionCmd(r *root) *cobra.Command {
	c := &version{
		root: r,
	}
	cmd := &cobra.Command{
		Use:     "version",
		Aliases: []string{"ver"},
		Short:   "Print client and server version information",
		Long:    "Print client and server versions for the provided context",
		RunE:    c.RunE,
	}

	pf := r.cmd.LocalFlags()

	pf.BoolVarP(&c.clientOnly, "client", "c", false, "Client version information only")
	pf.BoolVarP(&c.serverOnly, "server", "s", false, "Server version information only")

	return cmd
}

func (c *version) RunE(cmd *cobra.Command, _ []string) error {
	c.conf.Finalize()

	return c.retry(func() error {
		ver, err := c.client.Version(cmd.Context())
		if err != nil {
			return err
		}

		data := struct {
			ClientVersion string
			GoVersion     string
			Server        *kivik.Version
		}{
			ClientVersion: v.Version,
			GoVersion:     fmt.Sprintf("%s, %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH),
			Server:        ver,
		}

		format := `kubectl Client Version: {{ .ClientVersion }}, {{ .GoVersion }}
CouchDB Server Version: {{ .Server.Version }}, {{ .Server.Vendor }}`
		result := output.TemplateReader(format, data, bytes.NewReader(ver.RawResponse))
		return c.fmt.Output(result)
	})
}
