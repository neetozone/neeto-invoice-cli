package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestReportingCommandsRegistered(t *testing.T) {
	checks := []struct {
		parent *cobra.Command
		subs   []string
	}{
		{monthlyPtosCmd, []string{"list", "update-earned"}},
		{forcedPtosCmd, []string{"list", "create"}},
		{reportsCmd, []string{"payroll-summary", "missing-entries", "timesheet-summary"}},
	}

	for _, check := range checks {
		present := map[string]bool{}
		for _, sub := range check.parent.Commands() {
			present[sub.Name()] = true
		}
		for _, want := range check.subs {
			if !present[want] {
				t.Errorf("%s: subcommand %q not registered", check.parent.Name(), want)
			}
		}
	}
}

func TestDiscoveryListCommandsRegistered(t *testing.T) {
	checks := []struct {
		parent *cobra.Command
		sub    string
	}{
		{clientsCmd, "list"},
		{projectsCmd, "list"},
	}

	for _, check := range checks {
		found := false
		for _, sub := range check.parent.Commands() {
			if sub.Name() == check.sub {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: subcommand %q not registered", check.parent.Name(), check.sub)
		}
	}
}
