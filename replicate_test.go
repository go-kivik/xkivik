package xkivik

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/flimzy/diff"
	"github.com/flimzy/testy"

	"github.com/go-kivik/kivik"
	"github.com/go-kivik/kivik/driver"
	"github.com/go-kivik/kivikmock"
)

func TestReplicate(t *testing.T) {
	type tt struct {
		mockT, mockS   *kivikmock.Client
		target, source *kivik.DB
		options        kivik.Options
		status         int
		err            string
		result         *ReplicationResult
	}
	tests := testy.NewTable()
	tests.Add("changes error", func(t *testing.T) interface{} {
		source, mock := kivikmock.NewT(t)
		db := mock.NewDB()
		mock.ExpectDB().WillReturn(db)
		db.ExpectChanges().WillReturnError(errors.New("changes err"))

		return tt{
			mockS:  mock,
			source: source.DB(context.TODO(), "src"),
			status: http.StatusInternalServerError,
			err:    "changes err",
			result: &ReplicationResult{},
		}
	})
	tests.Add("no changes", func(t *testing.T) interface{} {
		source, mock := kivikmock.NewT(t)
		db := mock.NewDB()
		mock.ExpectDB().WillReturn(db)
		db.ExpectChanges().WillReturn(kivikmock.NewChanges())

		return tt{
			mockS:  mock,
			source: source.DB(context.TODO(), "src"),
			result: &ReplicationResult{},
		}
	})
	tests.Add("up to date", func(t *testing.T) interface{} {
		source, smock := kivikmock.NewT(t)
		sdb := smock.NewDB()
		smock.ExpectDB().WillReturn(sdb)
		sdb.ExpectChanges().WillReturn(kivikmock.NewChanges().
			AddChange(&driver.Change{
				ID:      "foo",
				Changes: []string{"2-7051cbe5c8faecd085a3fa619e6e6337"},
				Seq:     "3-g1AAAAG3eJzLYWBg4MhgTmHgz8tPSTV0MDQy1zMAQsMcoARTIkOS_P___7MSGXAqSVIAkkn2IFUZzIkMuUAee5pRqnGiuXkKA2dpXkpqWmZeagpu_Q4g_fGEbEkAqaqH2sIItsXAyMjM2NgUUwdOU_JYgCRDA5ACGjQfn30QlQsgKvcjfGaQZmaUmmZClM8gZhyAmHGfsG0PICrBPmQC22ZqbGRqamyIqSsLAAArcXo",
			}))

		target, tmock := kivikmock.NewT(t)
		tdb := tmock.NewDB()
		tmock.ExpectDB().WillReturn(tdb)
		tdb.ExpectRevsDiff().
			WithRevLookup(map[string][]string{
				"foo": {"2-7051cbe5c8faecd085a3fa619e6e6337"},
			}).
			WillReturn(kivikmock.NewRows())

		return tt{
			mockS:  smock,
			mockT:  tmock,
			source: source.DB(context.TODO(), "src"),
			target: target.DB(context.TODO(), "tgt"),
			result: &ReplicationResult{},
		}
	})
	tests.Add("one update", func(t *testing.T) interface{} {
		source, smock := kivikmock.NewT(t)
		sdb := smock.NewDB()
		smock.ExpectDB().WillReturn(sdb)
		sdb.ExpectChanges().WillReturn(kivikmock.NewChanges().
			AddChange(&driver.Change{
				ID:      "foo",
				Changes: []string{"2-7051cbe5c8faecd085a3fa619e6e6337"},
				Seq:     "3-g1AAAAG3eJzLYWBg4MhgTmHgz8tPSTV0MDQy1zMAQsMcoARTIkOS_P___7MSGXAqSVIAkkn2IFUZzIkMuUAee5pRqnGiuXkKA2dpXkpqWmZeagpu_Q4g_fGEbEkAqaqH2sIItsXAyMjM2NgUUwdOU_JYgCRDA5ACGjQfn30QlQsgKvcjfGaQZmaUmmZClM8gZhyAmHGfsG0PICrBPmQC22ZqbGRqamyIqSsLAAArcXo",
			}))

		target, tmock := kivikmock.NewT(t)
		tdb := tmock.NewDB()
		tmock.ExpectDB().WillReturn(tdb)
		tdb.ExpectRevsDiff().
			WithRevLookup(map[string][]string{
				"foo": {"2-7051cbe5c8faecd085a3fa619e6e6337"},
			}).
			WillReturn(kivikmock.NewRows().
				AddRow(&driver.Row{
					ID:    "foo",
					Value: []byte(`{"missing":["2-7051cbe5c8faecd085a3fa619e6e6337"]}`),
				}))
		sdb.ExpectGet().
			WithDocID("foo").
			WithOptions(kivik.Options{
				"rev":  "2-7051cbe5c8faecd085a3fa619e6e6337",
				"revs": true,
			}).
			WillReturn(kivikmock.DocumentT(t, `{"_id":"foo","_rev":"2-7051cbe5c8faecd085a3fa619e6e6337","foo":"bar"}`))
		tdb.ExpectPut().
			WithDocID("foo").
			WithOptions(kivik.Options{
				"new_edits": false,
			}).
			WillReturn("2-7051cbe5c8faecd085a3fa619e6e6337")

		return tt{
			mockS:  smock,
			mockT:  tmock,
			source: source.DB(context.TODO(), "src"),
			target: target.DB(context.TODO(), "tgt"),
			result: &ReplicationResult{
				DocsRead:       1,
				DocsWritten:    1,
				MissingChecked: 1,
				MissingFound:   1,
			},
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		result, err := Replicate(context.TODO(), tt.target, tt.source, tt.options)
		testy.StatusError(t, tt.err, tt.status, err)
		if tt.mockT != nil {
			testy.Error(t, "", tt.mockT.ExpectationsWereMet())
		}
		if tt.mockS != nil {
			testy.Error(t, "", tt.mockS.ExpectationsWereMet())
		}
		result.StartTime = time.Time{}
		result.EndTime = time.Time{}
		if d := diff.AsJSON(tt.result, result); d != nil {
			t.Error(d)
		}
	})
}
