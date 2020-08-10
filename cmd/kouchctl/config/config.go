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
	"net/url"

	"gopkg.in/yaml.v3"
)

// Context represents a complete, or partial CouchDB DSN context.
type Context struct {
	Name     string `yaml:"name"`
	Scheme   string `yaml:"scheme"`
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

// UnmarshalYAML handles parsing of a Context from YAML input.
func (c *Context) UnmarshalYAML(v *yaml.Node) error {
	dsn := struct {
		Name string `yaml:"name"`
		DSN  string `yaml:"dsn"`
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
		Name:     dsn.Name,
		Scheme:   uri.Scheme,
		Host:     uri.Host,
		User:     user,
		Password: password,
		Database: uri.Path,
	}
	return nil
}
