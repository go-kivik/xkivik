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
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type root struct {
	confFile string
	verbose  bool
	debug    bool
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context) {
	fmt.Println(os.Args)
	root := rootCmd()

	root.AddCommand(fooCmd())

	if err := root.ExecuteContext(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	r := &root{}

	cmd := &cobra.Command{
		Use:   "kouchctl",
		Short: "kouchctl facilitates controlling CouchDB instances",
		Long:  `This tool makes it easier to administrate and interact with CouchDB's HTTP API`,
		RunE:  r.RunE,
	}

	cmd.PersistentFlags().StringVar(&r.confFile, "kouchconfig", "~/.kouchctl/config", "Path to kouchconfig file to use for CLI requests")

	return cmd
}

func (r *root) RunE(cmd *cobra.Command, args []string) error {
	return nil
}
