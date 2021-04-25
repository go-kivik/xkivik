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

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/errors"
)

type postReplicate struct {
	*root
	source, target string
}

func postReplicateCmd(r *root) *cobra.Command {
	c := &postReplicate{
		root: r,
	}
	cmd := &cobra.Command{
		Use:     "replicate [dsn]/[database]",
		Aliases: []string{"rep"},
		Short:   "Replicate a database",
		RunE:    c.RunE,
	}

	pf := cmd.PersistentFlags()
	pf.StringVarP(&c.source, "source", "s", "", "The source DSN")
	pf.StringVarP(&c.target, "target", "t", "", "The target DSN")

	return cmd
}

func (c *postReplicate) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}

	if c.source == "" && c.target == "" {
		return errors.Code(errors.ErrUsage, "explicit source or target required")
	}

	c.log.Debugf("[post] Will replicate %s to %s", c.source, c.target)
	return c.retry(func() error {
		_, err := client.Replicate(cmd.Context(), c.target, c.source, c.opts())
		if err != nil {
			return err
		}
		return c.fmt.OK()
	})
}
