package xkivik

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"sync"
	"time"

	"gitlab.com/flimzy/parallel"

	"github.com/go-kivik/kivik"
)

func mergeOptions(otherOpts ...kivik.Options) kivik.Options {
	if len(otherOpts) == 0 {
		return nil
	}
	options := make(kivik.Options)
	for _, opts := range otherOpts {
		for k, v := range opts {
			options[k] = v
		}
	}
	if len(options) == 0 {
		return nil
	}
	return options
}

// ReplicationResult represents the result of a replication.
type ReplicationResult struct {
	DocWriteFailures int       `json:"doc_write_failures"`
	DocsRead         int       `json:"docs_read"`
	DocsWritten      int       `json:"docs_written"`
	EndTime          time.Time `json:"end_time"`
	MissingChecked   int       `json:"missing_checked"`
	MissingFound     int       `json:"missing_found"`
	StartTime        time.Time `json:"start_time"`
}

type resultWrapper struct {
	*ReplicationResult
	mu sync.Mutex
}

func (r *resultWrapper) read() {
	r.mu.Lock()
	r.DocsRead++
	r.mu.Unlock()
}

func (r *resultWrapper) missingChecked() {
	r.mu.Lock()
	r.MissingChecked++
	r.mu.Unlock()
}

func (r *resultWrapper) missingFound() {
	r.mu.Lock()
	r.MissingFound++
	r.mu.Unlock()
}

func (r *resultWrapper) writeError() {
	r.mu.Lock()
	r.DocWriteFailures++
	r.mu.Unlock()
}

func (r *resultWrapper) write() {
	r.mu.Lock()
	r.DocsWritten++
	r.mu.Unlock()
}

// Replicate performs a replication from source to target, using a limited
// version of the CouchDB replication protocol.
//
// The following options are supported:
//
//     filter (string) - The name of a filter function.
//     doc_ids (array of string) - Array of document IDs to be synchronized.
//     copy_security (bool) - When true, the security object is read from the
//                            source, and copied to the target, before the
//                            replication. Use with caution! The security object
//                            is not versioned, and will be unconditionally
//                            overwritten!
func Replicate(ctx context.Context, target, source *kivik.DB, options ...kivik.Options) (*ReplicationResult, error) {
	result := &resultWrapper{
		ReplicationResult: &ReplicationResult{
			StartTime: time.Now(),
		},
	}
	defer func() {
		result.EndTime = time.Now()
	}()
	opts := mergeOptions(options...)
	if _, sec := opts["copy_security"].(bool); sec {
		if err := copySecurity(ctx, target, source); err != nil {
			return result.ReplicationResult, err
		}
	}
	group := parallel.New(ctx)
	changes := make(chan *change)
	group.Go(func(ctx context.Context) error {
		defer close(changes)
		return readChanges(ctx, source, changes, opts)
	})

	diffs := make(chan *revDiff)
	group.Go(func(ctx context.Context) error {
		defer close(diffs)
		return readDiffs(ctx, target, changes, diffs)
	})

	docs := make(chan *Document)
	group.Go(func(ctx context.Context) error {
		defer close(docs)
		return readDocs(ctx, source, diffs, docs, result)
	})

	group.Go(func(ctx context.Context) error {
		return storeDocs(ctx, target, docs, result)
	})

	return result.ReplicationResult, group.Wait()
}

func copySecurity(ctx context.Context, target, source *kivik.DB) error {
	sec, err := source.Security(ctx)
	if err != nil {
		return fmt.Errorf("read security: %w", err)
	}
	if err := target.SetSecurity(ctx, sec); err != nil {
		return fmt.Errorf("set security: %w", err)
	}
	return nil
}

type change struct {
	ID      string
	Changes []string
}

func readChanges(ctx context.Context, db *kivik.DB, results chan<- *change, options kivik.Options) error {
	opts := kivik.Options{
		"feed":  "normal",
		"style": "all_docs",
	}
	for _, key := range []string{"filter", "doc_ids"} {
		if value, ok := options[key]; ok {
			opts[key] = value
		}
	}
	changes, err := db.Changes(ctx, opts)
	if err != nil {
		return fmt.Errorf("open changes feed: %w", err)
	}

	defer changes.Close() // nolint: errcheck
	for changes.Next() {
		ch := &change{
			ID:      changes.ID(),
			Changes: changes.Changes(),
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case results <- ch:
		}
	}
	if err := changes.Err(); err != nil {
		return fmt.Errorf("read changes feed: %w", err)
	}
	return nil
}

type revDiff struct {
	ID                string   `json:"-"`
	Missing           []string `json:"missing"`
	PossibleAncestors []string `json:"possible_ancestors"`
}

const rdBatchSize = 10

func readDiffs(ctx context.Context, db *kivik.DB, ch <-chan *change, results chan<- *revDiff) error {
	for {
		revMap := map[string][]string{}
		var change *change
		var ok bool
	loop:
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case change, ok = <-ch:
				if !ok {
					break loop
				}
				revMap[change.ID] = change.Changes
				if len(revMap) >= rdBatchSize {
					break loop
				}
			}
		}

		if len(revMap) == 0 {
			return nil
		}
		diffs, err := db.RevsDiff(ctx, revMap)
		if err != nil {
			return err
		}
		defer diffs.Close() // nolint: errcheck
		for diffs.Next() {
			var val revDiff
			if err := diffs.ScanValue(&val); err != nil {
				return err
			}
			val.ID = diffs.ID()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case results <- &val:
			}
		}
		if err := diffs.Err(); err != nil {
			return fmt.Errorf("read revs diffs: %w", err)
		}
	}
}

func readDocs(ctx context.Context, db *kivik.DB, diffs <-chan *revDiff, results chan<- *Document, result *resultWrapper) error {
	for {
		var rd *revDiff
		var ok bool
		select {
		case <-ctx.Done():
			return ctx.Err()
		case rd, ok = <-diffs:
			if !ok {
				return nil
			}
			for _, rev := range rd.Missing {
				result.missingChecked()
				d, err := readDoc(ctx, db, rd.ID, rev)
				if err != nil {
					return fmt.Errorf("read doc %s: %w", rd.ID, err)
				}
				result.read()
				result.missingFound()
				select {
				case <-ctx.Done():
					return ctx.Err()
				case results <- d:
				}
			}
		}
	}
}

func readDoc(ctx context.Context, db *kivik.DB, docID, rev string) (*Document, error) {
	doc := new(Document)
	row := db.Get(ctx, docID, kivik.Options{
		"rev":         rev,
		"revs":        true,
		"attachments": true,
	})
	if err := row.ScanDoc(&doc); err != nil {
		return nil, err
	}
	// TODO: It seems silly this is necessary... I need better attachment
	// handling in kivik.
	if row.Attachments != nil {
		for {
			att, err := row.Attachments.Next()
			if err != nil {
				if err != io.EOF {
					return nil, err
				}
				break
			}
			var content []byte
			switch att.ContentEncoding {
			case "":
				var err error
				content, err = ioutil.ReadAll(att.Content)
				if err != nil {
					return nil, err
				}
				if err := att.Content.Close(); err != nil {
					return nil, err
				}
			case "gzip":
				zr, err := gzip.NewReader(att.Content)
				if err != nil {
					return nil, err
				}
				content, err = ioutil.ReadAll(zr)
				if err != nil {
					return nil, err
				}
				if err := zr.Close(); err != nil {
					return nil, err
				}
				if err := att.Content.Close(); err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("Unknown encoding '%s' for attachment '%s'", att.ContentEncoding, att.Filename)
			}
			att.Stub = false
			att.Follows = false
			att.Content = ioutil.NopCloser(bytes.NewReader(content))
			doc.Attachments.Set(att.Filename, att)
		}
	}
	return doc, nil
}

func storeDocs(ctx context.Context, db *kivik.DB, docs <-chan *Document, result *resultWrapper) error {
	for doc := range docs {
		if _, err := db.Put(ctx, doc.ID, doc, kivik.Options{
			"new_edits": false,
		}); err != nil {
			result.writeError()
			return fmt.Errorf("store doc %s: %w", doc.ID, err)
		}
		result.write()
	}
	return nil
}
