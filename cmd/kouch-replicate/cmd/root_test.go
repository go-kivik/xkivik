package cmd

import (
	"net/url"
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestParseURL(t *testing.T) {
	tests := map[string]*url.URL{
		"https://admin:abc123@localhost:5984/foo": &url.URL{
			Scheme: "https",
			Host:   "localhost:5984",
			Path:   "/foo",
			User:   url.UserPassword("admin", "abc123"),
		},
		"http://example.com/foo": &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/foo",
		},
		"file:///foo.txt": &url.URL{
			Scheme: "file",
			Path:   "/foo.txt",
		},
		"/usr/local/db": &url.URL{
			Scheme: "file",
			Path:   "/usr/local/db",
		},
	}
	for addr, want := range tests {
		t.Run(addr, func(t *testing.T) {
			got, err := parseURL(addr)
			if err != nil {
				t.Fatal(err)
			}
			if d := testy.DiffInterface(want, got); d != nil {
				t.Error(d)
			}
		})
	}
}
