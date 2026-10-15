package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestReportsMissingEntriesUserEmailIsOptional(t *testing.T) {
	flag := reportsMissingEntriesCmd.Flags().Lookup("user-email")
	if flag == nil {
		t.Fatal("user-email flag missing on reports missing-entries")
	}
	if _, ok := flag.Annotations[cobra.BashCompOneRequiredFlag]; ok {
		t.Error("user-email is required on reports missing-entries, want it optional")
	}
}

func TestReportsMissingEntriesWithoutUserOmitsEmail(t *testing.T) {
	req, _, err := runAgainstServer(t, `{"count":0,"users":[]}`,
		"reports", "missing-entries", "--start-date", "2026-10-01", "--end-date", "2026-10-14")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if req.Query.Has("email") {
		t.Errorf("query has email = %q, want it omitted", req.Query.Get("email"))
	}
	if req.Query.Get("start_date") != "2026-10-01" || req.Query.Get("end_date") != "2026-10-14" {
		t.Errorf("query = %v, want both dates", req.Query)
	}
}

func TestReportsMissingEntriesWithUserSendsEmail(t *testing.T) {
	req, _, err := runAgainstServer(t, `{"count":0,"users":[]}`,
		"reports", "missing-entries", "--user-email", "sam@example.com", "--start-date", "2026-10-01", "--end-date", "2026-10-14")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if req.Query.Get("email") != "sam@example.com" {
		t.Errorf("query email = %q, want sam@example.com", req.Query.Get("email"))
	}
}
