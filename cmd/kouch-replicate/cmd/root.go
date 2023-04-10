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
	"net/http"
	"net/url"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	_ "github.com/go-kivik/couchdb/v4" // The CouchDB driver
	_ "github.com/go-kivik/fsdb/v4"    // The Filesystem driver
	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/xkivik/v4"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kouch-replicate",
	Short: "replicate to/from CouchDB databases",
	Long: `This tool implements a minimalistic version of the CouchDB replication
protocol.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cobra.NoArgs(cmd, args); err != nil {
			return err
		}
		ctx := context.TODO()
		sourceDSN, _ := cmd.Flags().GetString("source")
		targetDSN, _ := cmd.Flags().GetString("target")
		source, err := connect(sourceDSN)
		if err != nil {
			return fmt.Errorf("connect source '%s': %w", sourceDSN, err)
		}
		target, err := connect(targetDSN)
		if err != nil {
			return fmt.Errorf("connect target '%s': %w", targetDSN, err)
		}
		_, err = xkivik.Replicate(ctx, target, source)
		return err
	},
}

func connect(dsn string) (*kivik.DB, error) {
	uri, err := url.Parse(dsn)
	if err != nil {
		return nil, &kivik.Error{Status: http.StatusBadRequest, Err: err}
	}
	switch uri.Scheme {
	case "http", "https":
		idx := strings.LastIndex(dsn, "/")
		client, err := kivik.New("couch", dsn[:idx+1])
		if err != nil {
			return nil, err
		}
		db := client.DB(dsn[idx+1:])
		return db, db.Err()
	case "file", "":
		client, err := kivik.New("fs", "")
		if err != nil {
			// WTF?
			return nil, err
		}
		db := client.DB(dsn)
		return db, db.Err()
	default:
		return nil, &kivik.Error{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("unsupported URL scheme '%s'", uri.Scheme),
		}
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kouch-replicate.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringP("source", "s", "", "Replication source")
	must(rootCmd.MarkFlagRequired("source"))
	rootCmd.Flags().StringP("target", "t", "", "Replication target")
	must(rootCmd.MarkFlagRequired("target"))
	/*
				TODO:
				security
		        timeout
				create_target
				continuous
				filter
				doc_ids
	*/
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".kouch-replicate" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".kouch-replicate")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
