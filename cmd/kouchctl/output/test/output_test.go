package output

import (
	"testing"

	"github.com/spf13/pflag"

	"gitlab.com/flimzy/testy"

	// Formats
	"github.com/go-kivik/xkivik/v4/cmd/kouchctl/output"
	_ "github.com/go-kivik/xkivik/v4/cmd/kouchctl/output/json"
)

func TestOutput(t *testing.T) {
	type tt struct {
		args []string
		obj  interface{}
		err  string
	}

	tests := testy.NewTable()
	tests.Add("defaults", tt{
		obj: map[string]string{"x": "y"},
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		fmt := output.New()
		flags := pflag.NewFlagSet("x", pflag.ContinueOnError)
		fmt.ConfigFlags(flags)
		if err := flags.ParseAll(tt.args, nil); err != nil {
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
	})
}
