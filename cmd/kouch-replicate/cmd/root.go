package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	_ "github.com/go-kivik/couchdb" // The CouchDB driver
	_ "github.com/go-kivik/fsdb"    // The Filesystem driver
	"github.com/go-kivik/kivik"
	"github.com/go-kivik/xkivik"
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
		source, err := connect(ctx, sourceDSN)
		if err != nil {
			return err
		}
		target, err := connect(ctx, targetDSN)
		if err != nil {
			return err
		}
		_, err = xkivik.Replicate(ctx, target, source)
		return err
	},
}

// parseURL parses absolute URLs. Relative URLs are assumed to be file://
func parseURL(addr string) (*url.URL, error) {
	for _, scheme := range []string{"http://", "https://", "file://"} {
		if strings.HasPrefix(addr, scheme) {
			return url.Parse(addr)
		}
	}
	return &url.URL{
		Scheme: "file",
		Path:   addr,
	}, nil
}

func connect(ctx context.Context, dsn string) (*kivik.DB, error) {
	idx := strings.LastIndex(dsn, "/")
	dbname := dsn[idx:]
	root := dsn[:idx]
	uri, err := parseURL(root)
	if err != nil {
		return nil, err
	}
	var client *kivik.Client
	switch uri.Scheme {
	case "http", "https":
		client, err = kivik.New("couch", uri.String())
	case "file":
		client, err = kivik.New("fs", uri.Path)
	default:
		return nil, errors.Errorf("Unsupported URL scheme '%s'", uri.Scheme)
	}
	if err != nil {
		return nil, err
	}
	return client.DB(ctx, dbname), nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
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
	rootCmd.MarkFlagRequired("source")
	rootCmd.Flags().StringP("target", "t", "", "Replication target")
	rootCmd.MarkFlagRequired("target")
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
