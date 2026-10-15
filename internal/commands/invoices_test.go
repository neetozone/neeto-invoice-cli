package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestInvoicesCreateRequiresNumber(t *testing.T) {
	flag := invoicesCreateCmd.Flags().Lookup("number")
	if flag == nil {
		t.Fatal("number flag missing on invoices create")
	}
	if _, ok := flag.Annotations[cobra.BashCompOneRequiredFlag]; !ok {
		t.Error("number flag is not required on invoices create")
	}
}

func TestInvoicesCreateFailsWithoutNumber(t *testing.T) {
	if err := invoicesCreateCmd.ParseFlags([]string{"--client", "c1", "--user-email", "sam@example.com"}); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}
	t.Cleanup(func() {
		for _, name := range []string{"client", "user-email"} {
			_ = invoicesCreateCmd.Flags().Set(name, "")
			invoicesCreateCmd.Flags().Lookup(name).Changed = false
		}
	})

	if err := invoicesCreateCmd.PreRunE(invoicesCreateCmd, nil); err != nil {
		t.Fatalf("PreRunE() error = %v", err)
	}

	err := invoicesCreateCmd.ValidateRequiredFlags()
	if err == nil {
		t.Fatal("ValidateRequiredFlags() = nil, want an error naming number")
	}
	if !strings.Contains(err.Error(), "number") {
		t.Errorf("ValidateRequiredFlags() error = %v, want it to name number", err)
	}
}

func TestInvoicePayloadFileFallsBackToDeprecatedDataFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().String("data", "", "")

	if err := cmd.ParseFlags([]string{"--data", "legacy.json"}); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}
	if got := invoicePayloadFile(cmd); got != "legacy.json" {
		t.Errorf("invoicePayloadFile() = %q, want legacy.json", got)
	}

	if err := cmd.Flags().Set("json-file", "current.json"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if got := invoicePayloadFile(cmd); got != "current.json" {
		t.Errorf("invoicePayloadFile() = %q, want current.json", got)
	}
}

func TestInvoicesCreateDeprecatedDataFlagSatisfiesNumber(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invoice.json")
	if err := os.WriteFile(path, []byte(`{"number":"INV-2","invoice_services":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}

	args := []string{"--data", path, "--client", "c1", "--user-email", "sam@example.com"}
	if err := invoicesCreateCmd.ParseFlags(args); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}
	t.Cleanup(func() {
		for _, name := range []string{"data", "json-file", "client", "user-email"} {
			_ = invoicesCreateCmd.Flags().Set(name, "")
			invoicesCreateCmd.Flags().Lookup(name).Changed = false
		}
		_ = invoicesCreateCmd.MarkFlagRequired("number")
	})

	if err := invoicesCreateCmd.PreRunE(invoicesCreateCmd, nil); err != nil {
		t.Fatalf("PreRunE() error = %v", err)
	}
	if err := invoicesCreateCmd.ValidateRequiredFlags(); err != nil {
		t.Errorf("ValidateRequiredFlags() error = %v, want the JSON payload to satisfy --number", err)
	}
}

func TestInvoicesCreateAcceptsNumberFromJSONFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invoice.json")
	if err := os.WriteFile(path, []byte(`{"number":"INV-1","invoice_services":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{Use: "create"}
	cmd.Flags().String("json-file", "", "")
	cmd.Flags().String("number", "", "")
	if err := cmd.MarkFlagRequired("number"); err != nil {
		t.Fatalf("MarkFlagRequired() error = %v", err)
	}
	allowJSONFileToSatisfyRequiredFlags(cmd)
	if err := cmd.ParseFlags([]string{"--json-file", path}); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	if err := cmd.PreRunE(cmd, nil); err != nil {
		t.Fatalf("PreRunE() error = %v", err)
	}
	if _, ok := cmd.Flags().Lookup("number").Annotations[cobra.BashCompOneRequiredFlag]; ok {
		t.Error("number should no longer be required once the JSON file supplies it")
	}
}

func TestInvoicesListSendsFilters(t *testing.T) {
	req, _, err := runAgainstServer(t, `{"invoices":[],"total_count":0,"page":1,"page_size":30,"total_pages":1}`,
		"invoices", "list", "--client", "abc123", "--status", "generated", "--overdue",
		"--start-date", "2026-08-01", "--end-date", "2026-08-31", "--page", "2", "--page-size", "50")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if req.Method != "GET" || !strings.HasSuffix(req.Path, "/invoices") {
		t.Errorf("request = %s %s, want GET .../invoices", req.Method, req.Path)
	}
	want := map[string]string{
		"client_id": "abc123", "status": "generated", "overdue": "true",
		"start_date": "2026-08-01", "end_date": "2026-08-31",
	}
	for key, value := range want {
		if got := req.Query.Get(key); got != value {
			t.Errorf("query %s = %q, want %q", key, got, value)
		}
	}
	if req.Query.Get("page") != "2" || req.Query.Get("page_size") != "50" {
		t.Errorf("pagination query = %v, want page=2 page_size=50", req.Query)
	}
}

func TestInvoicesListOmitsUnsetFilters(t *testing.T) {
	req, _, err := runAgainstServer(t, `{"invoices":[]}`, "invoices", "list")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	for _, key := range []string{"client_id", "status", "overdue", "start_date", "end_date"} {
		if req.Query.Has(key) {
			t.Errorf("query has %s = %q, want it omitted", key, req.Query.Get(key))
		}
	}
}

func TestInvoicesSendPostsBody(t *testing.T) {
	req, _, err := runAgainstServer(t, `{"message":"sent","id":"i1","number":"INV-1","sent_date":"2026-10-15","recipient_emails":["a@example.com"]}`,
		"invoices", "send", "INV-2026-042", "--client", "abc123",
		"--recipient-id", "r1", "--recipient-id", "r2", "--subject", "Hello", "--message", "Please pay",
		"--attach-pdf", "--send-me-a-copy", "--user-email", "sam@example.com")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if req.Method != "POST" || !strings.HasSuffix(req.Path, "/clients/abc123/invoices/INV-2026-042/send") {
		t.Errorf("request = %s %s, want POST .../clients/abc123/invoices/INV-2026-042/send", req.Method, req.Path)
	}
	if got, _ := req.Body["recipient_ids"].([]any); len(got) != 2 || got[0] != "r1" || got[1] != "r2" {
		t.Errorf("recipient_ids = %v, want [r1 r2]", req.Body["recipient_ids"])
	}
	for key, value := range map[string]any{
		"subject": "Hello", "message": "Please pay", "attach_pdf_invoice": true,
		"send_me_a_copy": true, "email": "sam@example.com",
	} {
		if req.Body[key] != value {
			t.Errorf("body %s = %v, want %v", key, req.Body[key], value)
		}
	}
}

func TestInvoicesSendOmitsUnsetOptions(t *testing.T) {
	req, _, err := runAgainstServer(t, `{"message":"sent"}`, "invoices", "send", "INV-1", "--client", "abc123")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if len(req.Body) != 0 {
		t.Errorf("body = %v, want it empty", req.Body)
	}
}

func TestInvoicesSendRequiresClient(t *testing.T) {
	_, _, err := runAgainstServer(t, `{}`, "invoices", "send", "INV-1")
	if err == nil || !strings.Contains(err.Error(), "client") {
		t.Errorf("Execute() error = %v, want an error naming client", err)
	}
}

func TestInvoicesCreateSendsTimeEntryIDs(t *testing.T) {
	req, _, err := runAgainstServer(t, `{"notice_code":"thumbs_up","invoice":{}}`,
		"invoices", "create", "--client", "abc123", "--user-email", "sam@example.com", "--number", "INV-9",
		"--time-entry-id", "t1", "--time-entry-id", "t2", "--issue-date", "2026-10-15")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if got, _ := req.Body["time_entry_ids"].([]any); len(got) != 2 || got[0] != "t1" || got[1] != "t2" {
		t.Errorf("time_entry_ids = %v, want [t1 t2]", req.Body["time_entry_ids"])
	}
	if req.Body["issue_date"] != "2026-10-15" {
		t.Errorf("issue_date = %v, want 2026-10-15", req.Body["issue_date"])
	}
}

func TestInvoicesCreateDateHelpUsesISO(t *testing.T) {
	for _, name := range []string{"issue-date", "due-date"} {
		if usage := invoicesCreateCmd.Flags().Lookup(name).Usage; !strings.Contains(usage, "YYYY-MM-DD") {
			t.Errorf("%s usage = %q, want YYYY-MM-DD", name, usage)
		}
	}
}
