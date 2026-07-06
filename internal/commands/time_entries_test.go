package commands

import "testing"

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
