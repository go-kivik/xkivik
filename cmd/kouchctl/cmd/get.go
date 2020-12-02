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
)

type get struct {
	att, doc, db, ver *cobra.Command
	*root
}

func getCmd(r *root) *cobra.Command {
	g := &get{
		root: r,
		att:  getAttachmentCmd(r),
		doc:  getDocCmd(r),
		db:   getDBCmd(r),
		ver:  getVersionCmd(r),
	}
	cmd := &cobra.Command{
		Use:   "get [command]",
		Short: "Get a resource",
		Long:  `Fetch a resource described by the URL`,
		RunE:  g.RunE,
	}

	cmd.AddCommand(g.att)
	cmd.AddCommand(g.doc)
	cmd.AddCommand(g.db)
	cmd.AddCommand(g.ver)

	return cmd
}

func (g *get) RunE(cmd *cobra.Command, args []string) error {
	if g.conf.HasDoc() {
		return g.doc.RunE(cmd, args)
	}
	if g.conf.HasDB() {
		return g.db.RunE(cmd, args)
	}
	return g.ver.RunE(cmd, args)
}
