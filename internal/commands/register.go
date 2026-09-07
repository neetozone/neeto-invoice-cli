package commands

import (
	"encoding/json"
	"net/url"

	"github.com/neetozone/neeto-cli-commons/cli"
	"github.com/neetozone/neeto-cli-commons/client"
	"github.com/neetozone/neeto-cli-commons/output"
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

func printList(data json.RawMessage, resourceKey string, breadcrumbs []output.Breadcrumb) {
	app.PrintList(data, resourceKey, breadcrumbs)
}

func printResource(data json.RawMessage, breadcrumbs []output.Breadcrumb) {
	app.PrintResource(data, breadcrumbs)
}

func printActionResult(data json.RawMessage, breadcrumbs []output.Breadcrumb) {
	app.PrintActionResult(data, breadcrumbs)
}

func paginationParams(cmd *cobra.Command) url.Values { return app.PaginationParams(cmd) }

func addPaginationFlags(cmd *cobra.Command) { cli.AddPaginationFlags(0, cmd) }

func readJSONFile(path string) (map[string]any, error) { return cli.ReadJSONFile(path) }

func printMessage(msg string) { app.PrintMessage(msg) }

type inlinePaginationKeys struct {
	TotalRecords string
	TotalPages   string
	CurrentPage  string
	PageSize     string
}

func printListWithInlinePagination(data json.RawMessage, resourceKey string, keys inlinePaginationKeys, breadcrumbs []output.Breadcrumb) {
	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(data, &parsed); err != nil {
		app.Printer.Print(data, breadcrumbs)
		return
	}

	items, hasItems := parsed[resourceKey]
	if !hasItems {
		app.Printer.Print(data, breadcrumbs)
		return
	}

	pagination := client.Pagination{
		TotalRecords:      intFromRaw(parsed[keys.TotalRecords]),
		TotalPages:        intFromRaw(parsed[keys.TotalPages]),
		CurrentPageNumber: intFromRaw(parsed[keys.CurrentPage]),
		PageSize:          intFromRaw(parsed[keys.PageSize]),
	}

	paginationJSON, err := json.Marshal(pagination)
	if err != nil {
		app.Printer.Print(items, breadcrumbs)
		return
	}

	app.Printer.PrintWithPagination(items, paginationJSON, breadcrumbs)
}

func intFromRaw(raw json.RawMessage) int {
	var n int
	_ = json.Unmarshal(raw, &n)
	return n
}

func thumbsUpNotice(data json.RawMessage) (string, bool) {
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

func printCreateResult(data json.RawMessage, breadcrumbs []output.Breadcrumb) {
	if msg, ok := thumbsUpNotice(data); ok && !app.Printer.Toon && !app.Printer.UseJSON() {
		printMessage(msg)
		return
	}
	printActionResult(data, breadcrumbs)
}
