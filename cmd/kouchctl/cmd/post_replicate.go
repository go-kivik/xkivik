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
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/errors"
)

type postReplicate struct {
	*root
	source, target                   string
	cancel, continuous, createTarget bool
	docIDs                           []string
}

func postReplicateCmd(r *root) *cobra.Command {
	c := &postReplicate{
		root: r,
	}
	cmd := &cobra.Command{
		Use:     "replicate [dsn]",
		Aliases: []string{"rep"},
		Short:   "Replicate a database",
		Long:    "Creates a remotely-managed replication between source and target",
		RunE:    c.RunE,
	}

	pf := cmd.PersistentFlags()
	pf.StringVarP(&c.source, "source", "s", "", "The source DSN. String or JSON object")
	pf.StringVarP(&c.target, "target", "t", "", "The target DSN. String or JSON object")
	pf.BoolVar(&c.cancel, "cancel", false, "Cancel the replecation")
	pf.BoolVar(&c.continuous, "continuous", false, "Configure the replication to be continuous")
	pf.BoolVar(&c.createTarget, "create-target", false, "Creates the target database.")
	pf.StringSliceVar(&c.docIDs, "doc-id", nil, "Document IDs to be synchronized")

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
	opts := c.opts()
	if c.cancel {
		opts["cancel"] = c.cancel
	}
	if c.continuous {
		opts["continuous"] = c.continuous
	}
	if c.createTarget {
		opts["create_target"] = c.createTarget
	}
	if len(c.docIDs) > 0 {
		opts["doc_ids"] = c.docIDs
	}
	var source, target map[string]interface{}
	if err := json.Unmarshal([]byte(c.source), &source); err == nil {
		c.source = ""
		opts["source"] = source
	}
	if err := json.Unmarshal([]byte(c.target), &target); err == nil {
		c.target = ""
		opts["target"] = target
	}

	c.log.Debugf("[post] Will replicate %s to %s", c.source, c.target)
	return c.retry(func() error {
		_, err := client.Replicate(cmd.Context(), c.target, c.source, opts)
		if err != nil {
			return err
		}
		return c.fmt.OK()
	})
}
