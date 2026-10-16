package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/neetozone/neeto-cli-commons/auth"
	"github.com/neetozone/neeto-cli-commons/cli"
	"github.com/neetozone/neeto-cli-commons/config"
	product "github.com/neetozone/neeto-invoice-cli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestMain(m *testing.M) {
	cfg, err := config.Parse(product.ConfigYAML)
	if err != nil {
		panic(err)
	}
	Register(cli.New(*cfg))
	os.Exit(m.Run())
}

type recordedRequest struct {
	Method      string
	Path        string
	EscapedPath string
	Query       url.Values
	Body        map[string]any
}

func runAgainstServer(t *testing.T, response string, args ...string) (recordedRequest, string, error) {
	t.Helper()

	var recorded recordedRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorded = recordedRequest{Method: r.Method, Path: r.URL.Path, EscapedPath: r.URL.EscapedPath(), Query: r.URL.Query()}
		_ = json.NewDecoder(r.Body).Decode(&recorded.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)

	t.Setenv("HOME", t.TempDir())
	t.Setenv(app.Product.BaseURLEnvVar(), server.URL)
	if err := app.Auth.SaveStore(&auth.Store{Credentials: []auth.Credentials{{Subdomain: "acme", Email: "sam@example.com", SessionToken: "token"}}}); err != nil {
		t.Fatalf("SaveStore() error = %v", err)
	}

	out := &bytes.Buffer{}
	app.Printer.Out = out
	app.Printer.Err = out
	app.Root().SetOut(out)
	app.Root().SetErr(out)
	t.Cleanup(func() { resetFlags(app.Root()) })

	app.Root().SetArgs(append([]string{"--json"}, args...))
	err := app.Root().Execute()
	return recorded, out.String(), err
}

func resetFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if slice, ok := f.Value.(pflag.SliceValue); ok {
			_ = slice.Replace(nil)
		} else {
			_ = f.Value.Set(f.DefValue)
		}
		f.Changed = false
	})
	for _, sub := range cmd.Commands() {
		resetFlags(sub)
	}
}
