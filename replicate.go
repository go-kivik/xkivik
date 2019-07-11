package xkivik

import (
	"context"
	"encoding/json"
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
		return readChanges(ctx, source, changes)
	})

	diffs := make(chan *revDiff)
	group.Go(func(ctx context.Context) error {
		defer close(diffs)
		return readDiffs(ctx, target, changes, diffs)
	})

	docs := make(chan *doc)
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
		return err
	}
	return target.SetSecurity(ctx, sec)
}

type change struct {
	ID      string
	Changes []string
}

func readChanges(ctx context.Context, db *kivik.DB, results chan<- *change) error {
	changes, err := db.Changes(ctx, kivik.Options{"feed": "normal", "style": "all_docs"})
	if err != nil {
		return err
	}

	defer changes.Close() // nolint: errcheck
	for changes.Next() {
		results <- &change{
			ID:      changes.ID(),
			Changes: changes.Changes(),
		}
	}
	return changes.Err()
}

type revDiff struct {
	ID                string   `json:"-"`
	Missing           []string `json:"missing"`
	PossibleAncestors []string `json:"possible_ancestors"`
}

func readDiffs(ctx context.Context, db *kivik.DB, ch <-chan *change, results chan<- *revDiff) error {
	revMap := map[string][]string{}
	for change := range ch {
		revMap[change.ID] = change.Changes
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
		results <- &val
	}
	return diffs.Err()
}

type doc struct {
	ID      string
	Rev     string
	Content json.RawMessage
}

func readDocs(ctx context.Context, db *kivik.DB, diffs <-chan *revDiff, results chan<- *doc, result *resultWrapper) error {
	for rd := range diffs {
		for _, rev := range rd.Missing {
			result.missingChecked()
			d, err := readDoc(ctx, db, rd.ID, rev)
			if err != nil {
				return err
			}
			result.read()
			result.missingFound()
			results <- d
		}
	}

	return nil
}

func readDoc(ctx context.Context, db *kivik.DB, docID, rev string) (*doc, error) {
	row := db.Get(ctx, docID, kivik.Options{"rev": rev})
	if row.Err != nil {
		return nil, row.Err
	}
	defer row.Body.Close() // nolint: errcheck
	body, err := ioutil.ReadAll(row.Body)
	if err != nil {
		return nil, err
	}
	return &doc{
		ID:      docID,
		Rev:     rev,
		Content: body,
	}, nil
}

func storeDocs(ctx context.Context, db *kivik.DB, docs <-chan *doc, result *resultWrapper) error {
	for doc := range docs {
		if _, err := db.Put(ctx, doc.ID, doc.Content, kivik.Options{"new_edits": false}); err != nil {
			result.writeError()
			return err
		}
		result.write()
	}
	return nil
}
