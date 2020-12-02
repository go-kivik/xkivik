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

type create struct {
	*root
}

func createCmd(r *root) *cobra.Command {
	c := &create{
		root: r,
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a resource",
		Long:  `Create the named resource`,
		RunE:  c.RunE,
	}

	return cmd
}

func (c *create) RunE(cmd *cobra.Command, args []string) error {
	_, err := c.client()
	if err != nil {
		return err
	}

	return errors.New("xxx")
}
