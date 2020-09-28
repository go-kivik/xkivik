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

type ping struct {
	log  log.Logger
	conf *config.Config
}

func pingCmd(lg log.Logger, conf *config.Config) *cobra.Command {
	c := &ping{
		log:  lg,
		conf: conf,
	}

	return &cobra.Command{
		Use:   "ping [dsn]",
		Short: "Ping a server",
		Long:  "Ping a server's /_up endpoint to determine availability to serve requests",
		RunE:  c.RunE,
	}
}

func (c *ping) RunE(cmd *cobra.Command, args []string) error {
	dsn, err := c.conf.ServerDSN()
	if err != nil {
		return err
	}
	c.log.Debugf("[ping] Will ping server: %q", dsn)
	client, err := kivik.New("couch", dsn)
	if err != nil {
		return err
	}
	success, err := client.Ping(cmd.Context())
	if success {
		c.log.Info("[ping] Server is up")
	} else {
		c.log.Info("[ping] Server down")
	}
	return err
}