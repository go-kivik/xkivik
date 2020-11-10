// Package couchdb provides a (near) drop-in replacement for
// github.com/fjl/go-couchdb which is backed by Kivik. It was built to
// facilitate the transition from go-couchdb to Kivik.
package couchdb

import (
	"context"
	"io"
	"net/http"

	"github.com/go-kivik/couchdb/v4" // CouchDB driver
	"github.com/go-kivik/kivik/v4"
)

// Client represents a remote CouchDB server.
type Client struct {
	*kivik.Client
}

var _ clientIface = &Client{}

// NewClient creates a new client object.
//
// If dsn contains credentials, the client will authenticate using HTTP Basic
// Authentication. If rawurl has a query string, it is ignored.
//
// The second argument can be nil to use http.Transport, which should be good
// enough in most cases.
func NewClient(dsn string, rt http.RoundTripper) (*Client, error) {
	c, err := kivik.New("couch", dsn)
	if err != nil {
		return nil, err
	}
	if rt != nil {
		if err := c.Authenticate(context.Background(), couchdb.SetTransport(rt)); err != nil {
			return nil, err
		}
	}
	return &Client{Client: c}, nil
}

// KivikClient returns a new Client instance, backed by the existing
// *kivik.Client.
func KivikClient(c *kivik.Client) *Client {
	return &Client{Client: c}
}

// AllDBs returns the names of all existing databases.
func (c *Client) AllDBs() ([]string, error) {
	return c.Client.AllDBs(context.Background())
}

// CreateDB creates a new database.
func (c *Client) CreateDB(name string) (*DB, error) {
	err := c.Client.CreateDB(context.Background(), name)
	return c.DB(name), err
}

// DB creates a database object.
func (c *Client) DB(name string) *DB {
	return &DB{
		DB: c.Client.DB(name),
	}
}

// DBUpdates opens the _db_updates feed.
func (c *Client) DBUpdates(options Options) (*DBUpdatesFeed, error) {
	updates, err := c.Client.DBUpdates(context.Background(), options)
	if err != nil {
		return nil, err
	}
	return &DBUpdatesFeed{
		DBUpdates: updates,
	}, nil
}

// DeleteDB deletes an existing database.
func (c *Client) DeleteDB(name string) error {
	return c.DestroyDB(context.Background(), name)
}

// EnsureDB ensures that a database with the given name exists.
func (c *Client) EnsureDB(name string) (*DB, error) {
	db, err := c.CreateDB(name)
	if err != nil && kivik.StatusCode(err) != http.StatusPreconditionFailed {
		return nil, err
	}
	return db, nil
}

// Ping can be used to check whether a server is alive.
func (c *Client) Ping() error {
	ok, err := c.Client.Ping(context.Background())
	if err != nil {
		return err
	}
	if !ok {
		return &kivik.Error{HTTPStatus: http.StatusNotFound, Message: "server down"}
	}
	return nil
}

// SetAuth sets the authentication mechanism used by the client.
func (c *Client) SetAuth(a Auth) {
	_ = c.Authenticate(context.Background(), a)
}

// URL returns the URL prefix of the server.
func (c *Client) URL() string {
	return c.DSN()
}

// Options represents CouchDB query string parameters.
type Options = kivik.Options

type DBUpdatesFeed struct {
	Event string      `json:"type"`    // "created" | "updated" | "deleted"
	DB    string      `json:"db_name"` // Event database name
	Seq   interface{} `json:"seq"`     // DB update sequence of the event.
	OK    bool        `json:"ok"`      // Event operation status (deprecated)

	// DBUpdates is the underlying Kivik DBUpdates iterator.
	*kivik.DBUpdates
}

var _ updatesIface = &DBUpdatesFeed{}

// Auth is implemented by HTTP authentication mechanisms.
type Auth interface{}

// BasicAuth returns an Auth that performs HTTP Basic Authentication.
func BasicAuth(username, password string) Auth {
	return couchdb.BasicAuth(username, password)
}

// ProxyAuth returns an Auth that performs CouchDB proxy authentication.
func ProxyAuth(username string, roles []string, secret string) Auth {
	return couchdb.ProxyAuth(username, secret, roles)
}

// Attachment represents a document attachment.
type Attachment struct {
	Name string    // Filename
	Type string    // MIME type of the Body
	MD5  []byte    // MD5 checksum of the Body
	Body io.Reader // The body itself
}

type (
	// Security represents database security objects.
	Security = kivik.Security
	// Members represents member lists in database security objects.
	Members = kivik.Members
)
