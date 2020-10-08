package json

import (
	"encoding/json"
	"io"

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/output"
)

func init() {
	output.Register("json", &format{})
}

type format struct{}

var _ output.Format = &format{}

func (format) Output(w io.Writer, r io.Reader) error {
	var obj interface{}
	if err := json.NewDecoder(r).Decode(&obj); err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "\t")
	return enc.Encode(obj)
}
