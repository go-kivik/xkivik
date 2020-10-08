package output

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"

	"gitlab.com/flimzy/testy"

	// Formats
	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/output"
	_ "github.com/go-kivik/xkivik/v4/cmd/kouchctl/output/json"
)

func TestOutput(t *testing.T) {
	type tt struct {
		args  []string
		obj   interface{}
		err   string
		check func()
	}

	tests := testy.NewTable()
	tests.Add("defaults", tt{
		obj: map[string]string{"x": "y"},
	})
	tests.Add("output file", func(t *testing.T) interface{} {
		var dir string
		t.Cleanup(testy.TempDir(t, &dir))
		path := filepath.Join(dir, "test.json")

		return tt{
			args: []string{"-o", path},
			obj:  map[string]string{"x": "y"},
			check: func() {
				buf, err := ioutil.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if d := testy.DiffAsJSON([]byte(`{"x":"y"}`), buf); d != nil {
					t.Error(d)
				}
			},
		}
	})
	tests.Add("overwrite fail", func(t *testing.T) interface{} {
		var dir string
		t.Cleanup(testy.TempDir(t, &dir))
		path := filepath.Join(dir, "test.json")
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = fmt.Fprintf(f, "asdf")
		_ = f.Close()

		return tt{
			args: []string{"-o", path},
			obj:  map[string]string{"x": "y"},
			err:  "open " + path + ": file exists",
		}
	})
	tests.Add("overwrite success", func(t *testing.T) interface{} {
		var dir string
		t.Cleanup(testy.TempDir(t, &dir))
		path := filepath.Join(dir, "test.json")
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = fmt.Fprintf(f, "asdf")
		_ = f.Close()

		return tt{
			args: []string{"-o", path, "-O"},
			obj:  map[string]string{"x": "y"},
			check: func() {
				buf, err := ioutil.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if d := testy.DiffAsJSON([]byte(`{"x":"y"}`), buf); d != nil {
					t.Error(d)
				}
			},
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		fmt := output.New()
		flags := pflag.NewFlagSet("x", pflag.ContinueOnError)
		fmt.ConfigFlags(flags)

		set := func(flag *pflag.Flag, value string) error {
			return flags.Set(flag.Name, value)
		}

		if err := flags.ParseAll(tt.args, set); err != nil {
			t.Fatal(err)
		}
		var err error
		stdout, stderr := testy.RedirIO(nil, func() {
			err = fmt.Output(tt.obj)
		})

		testy.Error(t, tt.err, err)
		if d := testy.DiffText(testy.Snapshot(t, "_stdout"), stdout); d != nil {
			t.Errorf("STDOUT: %s", d)
		}
		if d := testy.DiffText("", stderr); d != nil {
			t.Errorf("STDERR: %s", d)
		}
		if tt.check != nil {
			tt.check()
		}
	})
}
