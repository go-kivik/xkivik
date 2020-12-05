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

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/input"
)

type putDoc struct {
	*root
	*input.Input
}

func putDocCmd(p *put) *cobra.Command {
	c := &putDoc{
		root:  p.root,
		Input: p.Input,
	}
	cmd := &cobra.Command{
		Use:     "document [dsn]/[database]/[document]",
		Aliases: []string{"doc"},
		Short:   "Put a document",
		Long:    `Create or update the named document`,
		RunE:    c.RunE,
	}

	return cmd
}

func (c *putDoc) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}
	doc, err := c.JSONData()
	if err != nil {
		return err
	}
	db, docID, err := c.conf.DBDoc()
	if err != nil {
		return err
	}
	c.log.Debugf("[put] Will put document: %s/%s/%s", client.DSN(), db, docID)
	return c.retry(func() error {
		rev, err := client.DB(db).Put(cmd.Context(), docID, doc, c.opts())
		if err != nil {
			return err
		}
		return c.fmt.UpdateResult(docID, rev)
	})
}
