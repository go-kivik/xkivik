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
	"errors"
	"net/url"
	"path"
	"strings"

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/config"
	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/log"
	"github.com/spf13/cobra"
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
		Use:   "get",
		Short: "get a document",
		Long:  `Fetch a document with the HTTP GET verb`,
		RunE:  g.RunE,
	}
}

func (c *get) RunE(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("no document specified to get")
	}
	dsn, err := url.Parse(args[0])
	if err != nil {
		return err
	}
	if dsn.Host == "" {
		cx, err := c.conf.DSN()
		if err != nil {
			return err
		}
		cxDSN, _ := url.Parse(cx)
		dsn.Scheme = cxDSN.Scheme
		dsn.Host = cxDSN.Host
		if !strings.Contains(dsn.Path, "/") {
			dsn.Path = path.Join(cxDSN.Path, dsn.Path)
		}
	}
	c.log.Debugf("[get] Will fetch document: %q", dsn)
	return nil
}
