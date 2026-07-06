package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestTimeEntriesHasDeleteAndUpdate(t *testing.T) {
	want := map[string]bool{"delete": false, "update": false}
	for _, c := range timeEntriesCmd.Commands() {
		if _, ok := want[c.Name()]; ok {
			want[c.Name()] = true
		}
	}
	for name, found := range want {
		if !found {
			t.Errorf("time-entries subcommand %q not registered", name)
		}
	}
}

func TestTimeEntriesFlags(t *testing.T) {
	if timeEntriesCreateCmd.Flags().Lookup("hours") == nil {
		t.Fatal("hours flag missing on create")
	}
	if timeEntriesCreateCmd.Flags().Lookup("is-override") == nil {
		t.Error("is-override flag missing on create")
	}
	if timeEntriesListCmd.Flags().Lookup("user-email") == nil {
		t.Error("user-email flag missing on list")
	}
	if timeEntriesUpdateCmd.Flags().Lookup("notes") == nil {
		t.Error("notes flag missing on update")
	}
}

func newTestCreateCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("client", "", "")
	cmd.Flags().String("project", "", "")
	cmd.Flags().String("task-id", "", "")
	cmd.Flags().String("user-email", "", "")
	cmd.Flags().String("recorded-on", "", "")
	cmd.Flags().Float64("hours", 0, "")
	cmd.Flags().String("notes", "", "")
	cmd.Flags().Bool("is-override", false, "")
	return cmd
}

func TestTimeEntryCreateBodyOmitsHoursAndOverrideWhenUnset(t *testing.T) {
	cmd := newTestCreateCmd()
	_ = cmd.ParseFlags([]string{"--client", "c1", "--project", "p1", "--task-id", "t1", "--user-email", "a@b.com", "--recorded-on", "2026-06-15"})

	body := timeEntryCreateBody(cmd)
	if _, ok := body["hours"]; ok {
		t.Error("hours should be omitted when not provided")
	}
	if _, ok := body["is_override"]; ok {
		t.Error("is_override should be omitted when not set")
	}
	if body["client_id"] != "c1" {
		t.Errorf("client_id = %v, want c1", body["client_id"])
	}
}

func TestTimeEntryCreateBodyIncludesExplicitZeroHoursAndOverride(t *testing.T) {
	cmd := newTestCreateCmd()
	_ = cmd.ParseFlags([]string{"--hours", "0", "--is-override"})

	body := timeEntryCreateBody(cmd)
	if body["hours"] != 0.0 {
		t.Errorf("hours = %v, want 0", body["hours"])
	}
	if body["is_override"] != true {
		t.Error("is_override should be true when flag is set")
	}
}

func TestTimeEntryUpdateBodyOnlyIncludesChangedFields(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("notes", "", "")
	cmd.Flags().Float64("hours", 0, "")
	cmd.Flags().String("recorded-on", "", "")
	cmd.Flags().Bool("is-override", false, "")
	cmd.Flags().String("user-email", "", "")
	_ = cmd.ParseFlags([]string{"--notes", "updated"})

	body := timeEntryUpdateBody(cmd)
	if body["notes"] != "updated" {
		t.Errorf("notes = %v, want updated", body["notes"])
	}
	if _, ok := body["hours"]; ok {
		t.Error("hours should be omitted when not changed")
	}
}
