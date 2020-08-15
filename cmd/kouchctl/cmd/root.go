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

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/config"
	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/log"
)

type root struct {
	confFile string
	debug    bool
	log      log.Logger
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

	pf := cmd.PersistentFlags()

	pf.StringVar(&r.confFile, "kouchconfig", "~/.kouchctl/config", "Path to kouchconfig file to use for CLI requests")
	pf.BoolVarP(&r.debug, "debug", "d", false, "Enable debug output")

	return cmd
}

func (r *root) RunE(cmd *cobra.Command, args []string) error {
	r.log = log.New(cmd)
	r.log.Debug("Debug mode enabled")

	conf, err := config.New(r.confFile)
	if err != nil {
		return err
	}
	cx, err := conf.DSN()
	if err != nil {
		return err
	}
	r.log.Debugf("DSN: %s from %q", cx, conf.CurrentContext)

	return nil
}
