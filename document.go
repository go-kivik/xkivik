package xkivik

import (
	"encoding/json"

	"github.com/go-kivik/kivik"
)

// Document represents any CouchDB document.
type Document struct {
	ID          string                 `json:"_id"`
	Rev         string                 `json:"_rev"`
	Attachments *kivik.Attachments     `json:"_attachments,omitempty"`
	Data        map[string]interface{} `json:"-"`
}

func (d *Document) MarshalJSON() ([]byte, error) {
	var data []byte
	doc, err := json.Marshal(*d)
	if err != nil {
		return nil, err
	}
	if len(d.Data) > 0 {
		var err error
		if data, err = json.Marshal(d.Data); err != nil {
			return nil, err
		}
		doc[len(doc)-1] = ','
		return append(doc, data[1:]...), nil
	}
	return doc, nil
}

func (d *Document) UnmarshalJSON(p []byte) error {
	type internalDoc Document
	doc := &internalDoc{}
	if err := json.Unmarshal(p, &doc); err != nil {
		return err
	}
	data := make(map[string]interface{})
	if err := json.Unmarshal(p, &data); err != nil {
		return err
	}
	delete(data, "_id")
	delete(data, "_rev")
	delete(data, "_attachments")
	*d = Document(*doc)
	d.Data = data
	return nil
}