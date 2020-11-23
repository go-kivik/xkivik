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

type descrDoc struct {
	*root
}

func descrDocCmd(r *root) *cobra.Command {
	g := descrDoc{
		root: r,
	}
	return &cobra.Command{
		Use:     "document [dsn]/[database]/[document]",
		Aliases: []string{"doc"},
		Short:   "Describe a document",
		Long:    `Fetch document metadata with the HTTP HEAD verb`,
		RunE:    g.RunE,
	}
}

func (c *descrDoc) RunE(cmd *cobra.Command, _ []string) error {
	db, docID, err := c.conf.DBDoc()
	if err != nil {
		return err
	}
	c.log.Debugf("[get] Will fetch document: %s/%s/%s", c.client.DSN(), db, docID)

	type result struct {
		ID   string `json:"_id"`
		Rev  string `json:"_rev"`
		Size int64  `json:"-"`
	}
	return c.retry(func() error {
		size, rev, err := c.client.DB(db).GetMeta(cmd.Context(), docID, c.opts())
		if err != nil {
			return err
		}
		data := result{
			ID:   docID,
			Rev:  rev,
			Size: size,
		}

		format := `      ID: {{ .ID }}
Revision: {{ .Rev }}
    Size: {{ .Size }}`
		result := output.TemplateReader(format, data, output.JSONReader(data))
		return c.fmt.Output(result)
	})
}
