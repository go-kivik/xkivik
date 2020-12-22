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
	"github.com/spf13/cobra"

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/output"
)

type getConfig struct {
	*root
	node string
}

func getConfigCmd(r *root) *cobra.Command {
	c := &getConfig{
		root: r,
	}
	cmd := &cobra.Command{
		Use:     "config [dsn]",
		Aliases: []string{"alldbs"},
		Short:   "Get server config",
		RunE:    c.RunE,
	}

	pf := cmd.PersistentFlags()
	pf.StringVarP(&c.node, "node", "n", "_local", "Specify the node name to query")

	return cmd
}

func (c *getConfig) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}
	c.conf.Finalize()

	return c.retry(func() error {
		conf, err := client.Config(cmd.Context(), c.node)
		if err != nil {
			return err
		}

		result := output.JSONReader(conf)
		return c.fmt.Output(result)
	})
}
