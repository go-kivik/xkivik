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
	"net"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	_ "github.com/go-kivik/couchdb/v4" // CouchDB driver

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/config"
	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/errors"
	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/log"
)

type root struct {
	confFile string
	debug    bool
	fail     bool
	log      log.Logger
	conf     *config.Config
	cmd      *cobra.Command
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ctx context.Context) {
	fmt.Println(os.Args)
	lg := log.New()
	root := rootCmd(lg)
	os.Exit(root.execute(ctx))
}

func (r *root) execute(ctx context.Context) int {
	err := r.cmd.ExecuteContext(ctx)
	if err == nil {
		return 0
	}
	code := extractExitCode(err, r.fail)

	return code
}

func extractExitCode(err error, fail bool) int {
	if code := errors.InspectErrorCode(err); code != 0 {
		if code >= http.StatusBadRequest {
			if fail {
				return errors.ErrHTTPPageNotRetrieved
			}
			return 0
		}
		return code
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return errors.ErrFailedToConnect
	}

	// Any unhandled errors are assumed to be from Cobra, so return a "failed
	// to initialize" error
	return errors.ErrFailedToInitialize
}

func rootCmd(lg log.Logger) *root {
	r := &root{
		log: lg,
	}
	r.cmd = &cobra.Command{
		Use:               "kouchctl",
		Short:             "kouchctl facilitates controlling CouchDB instances",
		Long:              `This tool makes it easier to administrate and interact with CouchDB's HTTP API`,
		PersistentPreRunE: r.init,
		RunE:              r.RunE,
	}
	r.conf = config.New(func() {
		r.cmd.SilenceUsage = true
	})

	pf := r.cmd.PersistentFlags()

	pf.StringVar(&r.confFile, "kouchconfig", "~/.kouchctl/config", "Path to kouchconfig file to use for CLI requests")
	pf.BoolVarP(&r.debug, "debug", "d", false, "Enable debug output")
	pf.BoolVarP(&r.fail, "fail", "f", false, "Fail silently (no output at all) on server errors")

	r.cmd.AddCommand(getCmd(r.log, r.conf))
	r.cmd.AddCommand(pingCmd(r.log, r.conf))

	return r
}

func (r *root) init(cmd *cobra.Command, args []string) error {
	r.log.SetOut(cmd.OutOrStdout())
	r.log.SetErr(cmd.ErrOrStderr())
	r.log.SetDebug(r.debug)

	r.log.Debug("Debug mode enabled")

	if err := r.conf.Read(r.confFile, r.log); err != nil {
		return err
	}

	if len(args) > 0 {
		if err := r.conf.SetURL(args[0]); err != nil {
			return err
		}
	}

	return nil
}

func (r *root) RunE(cmd *cobra.Command, args []string) error {
	cx, err := r.conf.DSN()
	if err != nil {
		return err
	}
	r.log.Debugf("DSN: %s from %q", cx, r.conf.CurrentContext)

	return nil
}
