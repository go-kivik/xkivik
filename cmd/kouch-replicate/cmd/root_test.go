package cmd

import (
	"context"
	"net/http"
	"regexp"
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestConnect(t *testing.T) {
	type tt struct {
		dsn    string
		status int
		err    string
	}
	tests := testy.NewTable()
	tests.Add("invalid url", tt{
		dsn:    "http://%xxx",
		status: http.StatusBadRequest,
		err:    `parse "?http://%xxx"?: invalid URL escape "%xx"`,
	})
	tests.Add("valid http:// url", tt{
		dsn: "http://example.com/foo",
	})
	tests.Add("valid https:// url", tt{
		dsn: "https://example.com/bar",
	})
	tests.Add("valid file:// url", tt{
		dsn: "file:///foo/bar",
	})
	tests.Add("unsupported scheme", tt{
		dsn:    "ftp://webmaster@www.google.com/",
		status: http.StatusBadRequest,
		err:    "unsupported URL scheme 'ftp'",
	})
	tests.Add("file:// url with invalid dbname", tt{
		dsn:    "file:///foo/bar.baz",
		status: http.StatusBadRequest,
		err:    regexp.QuoteMeta("Name: 'bar.baz'. Only lowercase characters (a-z), digits (0-9), and any of the characters _, $, (, ), +, -, and / are allowed. Must begin with a letter."),
	})
	tests.Add("local absolute path", tt{
		dsn: "/foo/bar",
	})
	tests.Add("local relative path", tt{
		dsn: "foo/bar",
	})
	tests.Add("local dot path", tt{
		dsn: "./foo/bar",
	})
	tests.Add("dot", tt{
		dsn: ".",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		result, err := connect(context.TODO(), tt.dsn)
		testy.StatusErrorRE(t, tt.err, tt.status, err)
		if d := testy.DiffInterface(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}
