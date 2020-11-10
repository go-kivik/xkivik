package couchdb

import (
	"encoding/json"
	"fmt"

	"github.com/go-kivik/kivik/v4"
)

// ChangesFeed is an iterator for the _changes feed of a database.
type ChangesFeed struct {
	// DB is the database. Since all events in a _changes feed
	// belong to the same database, this field is always equivalent to the
	// database from the DB.Changes call that created the feed object
	DB *DB `json:"-"`

	// ID is the document ID of the current event.
	ID string `json:"id"`

	// Deleted is true when the event represents a deleted document.
	Deleted bool `json:"deleted"`

	// Seq is the database update sequence number of the current event.
	// This is usually a string, but may also be a number for couchdb 0.x servers.
	//
	// For poll-style feeds (feed modes "normal", "longpoll"), this is set to the
	// last_seq value sent by CouchDB after all feed rows have been read.
	Seq interface{} `json:"seq"`

	// Pending is the count of remaining items in the feed. This is set for poll-style
	// feeds (feed modes "normal", "longpoll") after the last element has been
	// processed.
	Pending int64 `json:"pending"`

	// Changes is the list of the document's leaf revisions.
	Changes []struct {
		Rev string `json:"rev"`
	} `json:"changes"`

	// The document. This is populated only if the feed option
	// "include_docs" is true.
	Doc json.RawMessage `json:"doc"`

	ch *kivik.Changes
}

var _ changesIface = &ChangesFeed{}

// ChangesRevs returns the rev list of the current result row.
func (c *ChangesFeed) ChangesRevs() []string {
	return c.ch.Changes()
}

func (c *ChangesFeed) Close() error {
	return c.ch.Close()
}

func (c *ChangesFeed) Err() error {
	return c.ch.Err()
}

func (c *ChangesFeed) Next() bool {
	fmt.Println("Next()")
	next := c.ch.Next()

	type change struct {
		Rev string `json:"rev"`
	}

	c.ID = c.ch.ID()
	c.Deleted = c.ch.Deleted()
	c.Seq = c.ch.Seq()
	c.Changes = make([]struct {
		Rev string `json:"rev"`
	}, 0, len(c.ch.Changes()))
	for _, rev := range c.ch.Changes() {
		c.Changes = append(c.Changes, change{Rev: rev})
	}

	_ = c.ch.ScanDoc(&c.Doc)

	if next {
		return true
	}

	c.Pending = c.ch.Pending()
	c.Seq = c.ch.LastSeq()
	return false
}
