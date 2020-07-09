package xkivik

import (
	"encoding/json"
	"os"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v3"
)

func TestDocumentMarshalJSON(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("no attachments", &Document{
		ID:  "foo",
		Rev: "1-xxx",
		Data: map[string]interface{}{
			"foo": "bar",
		},
	})
	tests.Add("attachment", func(t *testing.T) interface{} {
		f, err := os.Open("testdata/foo.txt")
		if err != nil {
			t.Fatal(err)
		}

		return &Document{
			ID:  "foo",
			Rev: "1-xxx",
			Attachments: &kivik.Attachments{
				"foo.txt": &kivik.Attachment{
					ContentType: "text/plain",
					Content:     f,
				},
			},
		}
	})

	tests.Run(t, func(t *testing.T, doc *Document) {
		result, err := json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		if d := testy.DiffAsJSON(&testy.File{Path: "testdata/" + testy.Stub(t)}, result); d != nil {
			t.Error(d)
		}
	})
}

func TestNormalDocUnmarshalJSON(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("no extra fields", `{
        "_id":"foo"
    }`)
	tests.Add("extra fields", `{
        "_id":"foo",
        "foo":"bar"
    }`)
	tests.Add("attachment stub", `{
        "_id":"foo",
        "foo":"bar",
        "_attachments":{
            "foo.txt":{
                "content_type":"text/plain",
                "stub":true
            }
        }
    }`)
	tests.Add("attachment", `{
        "_id":"foo",
        "foo":"bar",
        "_attachments":{
            "foo.txt":{
                "content_type":"text/plain",
                "data":"VGVzdGluZwo="
            }
        }
    }`)

	tests.Run(t, func(t *testing.T, in string) {
		result := new(Document)
		if err := json.Unmarshal([]byte(in), &result); err != nil {
			t.Fatal(err)
		}
		if d := testy.DiffAsJSON(&testy.File{Path: "testdata/" + testy.Stub(t)}, result); d != nil {
			t.Error(d)
		}
	})
}
