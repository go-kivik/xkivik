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

package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"
)

const envPrefix = "KOUCHCTL"

// Config is the full app configuration file.
type Config struct {
	Contexts       map[string]*Context `yaml:"contexts"`
	CurrentContext string              `yaml:"current-context"`
}

// Context represents a complete, or partial CouchDB DSN context.
type Context struct {
	Scheme   string `yaml:"scheme"`
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func (c *Context) DSN() string {
	dsn := &url.URL{
		Scheme: c.Scheme,
		Host:   c.Host,
		Path:   c.Database,
		User:   url.UserPassword(c.User, c.Password),
	}
	return dsn.String()
}

// UnmarshalYAML handles parsing of a Context from YAML input.
func (c *Context) UnmarshalYAML(v *yaml.Node) error {
	dsn := struct {
		DSN string `yaml:"dsn"`
	}{}
	if err := v.Decode(&dsn); err != nil {
		return err
	}
	if dsn.DSN == "" {
		type alias Context
		intl := alias{}
		err := v.Decode(&intl)
		*c = Context(intl)
		return err
	}
	uri, err := url.Parse(dsn.DSN)
	if err != nil {
		return err
	}
	var user, password string
	if u := uri.User; u != nil {
		user = u.Username()
		password, _ = u.Password()
	}
	*c = Context{
		Scheme:   uri.Scheme,
		Host:     uri.Host,
		User:     user,
		Password: password,
		Database: uri.Path,
	}
	return nil
}

// New returns app configuration.
//
// - Reads from filename
// - If DSN env variable is set, it's added as context called 'ENV' and made current
func New(filename string) (*Config, error) {
	cf, err := readYAML(filename)
	if err != nil {
		return nil, err
	}
	if dsn := os.Getenv(envPrefix + "_DSN"); dsn != "" {
		uri, err := url.Parse(dsn)
		if err != nil {
			return nil, err
		}
		var user, password string
		if u := uri.User; u != nil {
			user = u.Username()
			password, _ = u.Password()
		}
		cf.Contexts["ENV"] = &Context{
			Scheme:   uri.Scheme,
			Host:     uri.Host,
			User:     user,
			Password: password,
			Database: uri.Path,
		}
		cf.CurrentContext = "ENV"
	}
	return cf, nil
}

func readYAML(filename string) (*Config, error) {
	cf := &Config{
		Contexts: make(map[string]*Context),
	}
	if filename == "" {
		return cf, nil
	}
	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return cf, err
	}
	if err := yaml.NewDecoder(f).Decode(cf); err != nil {
		return nil, err
	}
	return cf, nil
}

func (c *Config) DSN() (string, error) {
	if c.CurrentContext == "" {
		if len(c.Contexts) == 1 {
			for _, cx := range c.Contexts {
				return cx.DSN(), nil
			}
		}
		return "", errors.New("no context specified")
	}
	cx, ok := c.Contexts[c.CurrentContext]
	if !ok {
		return "", fmt.Errorf("context %q not found", c.CurrentContext)
	}
	return cx.DSN(), nil
}
