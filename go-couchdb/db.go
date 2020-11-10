package couchdb

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/go-kivik/kivik/v4"
)

type DB struct {
	*kivik.DB
}

var _ dbIface = &DB{}

type row struct {
	Bookmark string          `json:"bookmark,omitempty"`
	ID       string          `json:"id"`
	Key      json.RawMessage `json:"key"`
	Value    json.RawMessage `json:"value"`
	Doc      json.RawMessage `json:"doc,omitempty"`
}

type rows struct {
	Offset    int64  `json:"offset"`
	Rows      []*row `json:"rows"`
	TotalRows int64  `json:"total_rows"`
	UpdateSeq string `json:"update_seq,omitempty"`
	Warning   string `json:"warning,omitempty"`
}

// AllDocs invokes the _all_docs view of a database.
func (d *DB) AllDocs(result interface{}, opts Options) error {
	rows, err := d.DB.AllDocs(context.Background(), opts)
	if err != nil {
		return err
	}
	return readRows(rows, result)
}

func readRows(r *kivik.Rows, i interface{}) error {
	result := &rows{
		Rows: []*row{},
	}
	var doc, key, value json.RawMessage
	for r.Next() {
		_ = r.ScanDoc(&doc)
		_ = r.ScanKey(&key)
		_ = r.ScanValue(&value)
		result.Rows = append(result.Rows, &row{
			Bookmark: r.Bookmark(),
			ID:       r.ID(),
			Key:      key,
			Value:    value,
			Doc:      doc,
		})
	}
	if err := r.Err(); err != nil {
		return err
	}
	result.Offset = r.Offset()
	result.TotalRows = r.TotalRows()
	result.UpdateSeq = r.UpdateSeq()
	result.Warning = r.Warning()
	intermediate, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return json.Unmarshal(intermediate, &i)
}

// Attachment retrieves an attachment.
func (d *DB) Attachment(docid, name, rev string) (*Attachment, error) {
	att, err := d.DB.GetAttachment(context.Background(), docid, name, kivik.Options{
		"rev": rev,
	})
	if err != nil {
		return nil, err
	}
	md5, _ := base64.StdEncoding.DecodeString(att.Digest)
	body := &bytes.Buffer{}
	_, _ = io.Copy(body, att.Content)
	_ = att.Content.Close()
	return &Attachment{
		Name: att.Filename,
		Type: att.ContentType,
		MD5:  md5,
		Body: body,
	}, nil
}

// AttachmentMeta requests attachment metadata.
func (d *DB) AttachmentMeta(docid, name, rev string) (*Attachment, error) {
	att, err := d.DB.GetAttachmentMeta(context.Background(), docid, name, kivik.Options{
		"rev": rev,
	})
	if err != nil {
		return nil, err
	}
	md5, _ := base64.StdEncoding.DecodeString(att.Digest)
	return &Attachment{
		Name: att.Filename,
		Type: att.ContentType,
		MD5:  md5,
	}, nil
}

func (d *DB) Changes(options Options) (*ChangesFeed, error) {
	changes, err := d.DB.Changes(context.Background(), options)
	if err != nil {
		return nil, err
	}
	// spew.Dump(changes)
	return &ChangesFeed{
		DB: d,
		ch: changes,
	}, nil
}

// Delete marks a document revision as deleted.
func (d *DB) Delete(id, rev string) (string, error) {
	return d.DB.Delete(context.Background(), id, rev)
}

// DeleteAttachment removes an attachment.
func (d *DB) DeleteAttachment(docid, name, rev string) (string, error) {
	return d.DB.DeleteAttachment(context.Background(), docid, rev, name)
}

// Get retrieves a document from the given database.
func (d *DB) Get(id string, doc interface{}, opts Options) error {
	return d.DB.Get(context.Background(), id, opts).ScanDoc(doc)
}

// Put stores a document into the given database.
func (d *DB) Put(id string, doc interface{}, rev string) (string, error) {
	return d.DB.Put(context.Background(), id, doc, kivik.Options{"rev": rev})
}

// PutAttachment creates or updates an attachment.
func (d *DB) PutAttachment(docid string, att *Attachment, rev string) (string, error) {
	return d.DB.PutAttachment(context.Background(), docid, rev, &kivik.Attachment{
		Filename:    att.Name,
		ContentType: att.Type,
		Content:     ioutil.NopCloser(att.Body),
	})
}

// PutSecurity sets the database security object.
func (d *DB) PutSecurity(secobj *Security) error {
	return d.DB.SetSecurity(context.Background(), secobj)
}

// Rev fetches the current revision of a document.
func (d *DB) Rev(id string) (string, error) {
	_, rev, err := d.DB.GetMeta(context.Background(), id)
	return rev, err
}

// Security retrieves the security object of a database.
func (d *DB) Security() (*Security, error) {
	return d.DB.Security(context.Background())
}

// View invokes a view.
func (d *DB) View(ddoc, view string, result interface{}, opts Options) error {
	rows, err := d.DB.Query(context.Background(), ddoc, view, opts)
	if err != nil {
		return err
	}
	return readRows(rows, result)
}
