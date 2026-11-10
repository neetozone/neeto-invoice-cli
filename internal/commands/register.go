package commands

import (
	"encoding/json/jsontext"
	json "encoding/json/v2"
	"net/url"

	"github.com/neetozone/neeto-cli-commons/v2/cli"
	"github.com/neetozone/neeto-cli-commons/v2/client"
	"github.com/neetozone/neeto-cli-commons/v2/output"
	"github.com/spf13/cobra"
)

var (
	app           *cli.App
	registrations []func(root *cobra.Command)
)

func register(fn func(root *cobra.Command)) {
	registrations = append(registrations, fn)
}

func Register(a *cli.App) {
	app = a
	for _, fn := range registrations {
		fn(a.Root())
	}
}

func getClient(cmd *cobra.Command) (*client.Client, error) { return app.Client(cmd) }

func printList(data jsontext.Value, resourceKey string, breadcrumbs []output.Breadcrumb) {
	app.PrintList(data, resourceKey, breadcrumbs)
}

func printResource(data jsontext.Value, breadcrumbs []output.Breadcrumb) {
	app.PrintResource(data, breadcrumbs)
}

func printActionResult(data jsontext.Value, breadcrumbs []output.Breadcrumb) {
	app.PrintActionResult(data, breadcrumbs)
}

func paginationParams(cmd *cobra.Command) url.Values { return app.PaginationParams(cmd) }

func addPaginationFlags(cmd *cobra.Command) { cli.AddPaginationFlags(0, cmd) }

func readJSONFile(path string) (map[string]any, error) { return cli.ReadJSONFile(path) }

func allowJSONFileToSatisfyRequiredFlags(cmd *cobra.Command) {
	cli.AllowJSONFileToSatisfyRequiredFlags(cmd)
}

func printMessage(msg string) { app.PrintMessage(msg) }

func thumbsUpNotice(data jsontext.Value) (string, bool) {
	var obj struct {
		Notice     string `json:"notice"`
		NoticeCode string `json:"notice_code"`
	}
	if err := json.Unmarshal(data, &obj); err != nil {
		return "", false
	}
	if obj.NoticeCode != "thumbs_up" {
		return "", false
	}
	if obj.Notice != "" {
		return obj.Notice, true
	}
	return "success", true
}

func printCreateResult(data jsontext.Value, breadcrumbs []output.Breadcrumb) {
	if msg, ok := thumbsUpNotice(data); ok && !app.Printer.Toon && !app.Printer.UseJSON() {
		printMessage(msg)
		return
	}
	printActionResult(data, breadcrumbs)
}
