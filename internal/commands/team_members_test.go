package commands

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestTeamMembersCreateEmailAcceptsCommaSeparatedValues(t *testing.T) {
	cmd := newTestTeamMembersCreateCmd()
	if err := cmd.ParseFlags([]string{"--email", "oliver@example.com,sam@example.com"}); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	emails, _ := cmd.Flags().GetStringSlice("email")
	if len(emails) != 2 || emails[0] != "oliver@example.com" || emails[1] != "sam@example.com" {
		t.Errorf("emails = %v, want both addresses split apart", emails)
	}
}

func TestTeamMembersCreateEmailStaysRepeatable(t *testing.T) {
	cmd := newTestTeamMembersCreateCmd()
	if err := cmd.ParseFlags([]string{"--email", "oliver@example.com", "--email", "sam@example.com"}); err != nil {
		t.Fatalf("ParseFlags() error = %v", err)
	}

	emails, _ := cmd.Flags().GetStringSlice("email")
	if len(emails) != 2 {
		t.Errorf("emails = %v, want two entries", emails)
	}
}

func TestTeamMembersCreateEmailTypeMatchesOtherRepeatableFlags(t *testing.T) {
	want := monthlyPtosUpdateEarnedCmd.Flags().Lookup("email").Value.Type()
	if got := teamMembersCreateCmd.Flags().Lookup("email").Value.Type(); got != want {
		t.Errorf("team-members create --email type = %q, want %q to match the other repeatable email flag", got, want)
	}
}

func newTestTeamMembersCreateCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().StringSlice("email", nil, "")
	return cmd
}
