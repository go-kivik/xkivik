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
	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/doc"
	"github.com/spf13/cobra"
)

type put struct {
	*root
	doc *doc.Doc
}

func putCmd(r *root) *cobra.Command {
	p := &put{
		root: r,
		doc:  doc.New(),
	}
	cmd := &cobra.Command{
		Use:   "put [dsn]/[database]/[document]",
		Short: "Put a document",
		Long:  `Update or create a named document`,
		RunE:  p.RunE,
	}

	p.doc.ConfigFlags(cmd.Flags())

	return cmd
}

func (c *put) RunE(cmd *cobra.Command, _ []string) error {
	doc, err := c.doc.Data()
	if err != nil {
		return err
	}
	db, docID, err := c.conf.DBDoc()
	if err != nil {
		return err
	}
	c.log.Debugf("[put] Will put document: %s/%s/%s", c.client.DSN(), db, docID)
	return c.retry(func() error {
		rev, err := c.client.DB(db).Put(cmd.Context(), docID, doc, c.opts())
		if err != nil {
			return err
		}
		c.log.Info(rev)
		return nil
	})
}
