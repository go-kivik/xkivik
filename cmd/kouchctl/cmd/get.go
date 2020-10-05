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

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/config"
	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/log"
)

type get struct {
	log  log.Logger
	conf *config.Config
}

func getCmd(lg log.Logger, conf *config.Config) *cobra.Command {
	g := &get{
		log:  lg,
		conf: conf,
	}
	return &cobra.Command{
		Use:   "get [dsn]",
		Short: "get a document",
		Long:  `Fetch a document with the HTTP GET verb`,
		RunE:  g.RunE,
	}
}

func (c *get) RunE(cmd *cobra.Command, _ []string) error {
	dsn, db, doc, err := c.conf.DSNDoc()
	if err != nil {
		return err
	}
	c.log.Debugf("[get] Will fetch document: %s%s/%s", dsn, db, doc)
	client, err := kivik.New("couch", dsn)
	if err != nil {
		return err
	}
	row := client.DB(db).Get(cmd.Context(), doc)
	if err := row.Err; err != nil {
		return err
	}
	return nil
}
