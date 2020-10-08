package json

import (
	"io"

	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/output"
)

func init() {
	output.Register("raw", &format{})
}

type format struct{}

var _ output.Format = &format{}

func (format) Output(w io.Writer, r io.Reader) error {
	_, err := io.Copy(w, r)
	return err
}
