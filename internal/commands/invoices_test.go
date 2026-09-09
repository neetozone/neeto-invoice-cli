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
