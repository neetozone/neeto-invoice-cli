package commands

import (
	"fmt"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var teamMembersCmd = &cobra.Command{
	Use:   "team-members",
	Short: "Manage team members",
}

var teamMembersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List team members",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := paginationParams(cmd)
		if email, _ := cmd.Flags().GetString("email"); email != "" {
			params.Set("email", email)
		}

		data, err := c.Get("/team-members", params)
		if err != nil {
			return err
		}

		printList(data, "team_members", []output.Breadcrumb{
			{Label: "Show", Command: "neetoinvoice team-members show <id>"},
		})
		return nil
	},
}

var teamMembersShowCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show a team member",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/team-members/%s", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "Update", Command: "neetoinvoice team-members update <id> --role <role>"},
		})
		return nil
	},
}

var teamMembersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Invite team members",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		emails, _ := cmd.Flags().GetStringArray("email")
		role, _ := cmd.Flags().GetString("role")
		sendInvite, _ := cmd.Flags().GetBool("send-invitation-email")
		invitedBy, _ := cmd.Flags().GetString("invited-by")

		body := map[string]interface{}{
			"emails":                emails,
			"organization_role":     role,
			"send_invitation_email": sendInvite,
			"invited_by":            invitedBy,
		}

		data, err := c.Post("/team-members", body)
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{
			{Label: "List", Command: "neetoinvoice team-members list"},
		})
		return nil
	},
}

var teamMembersUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a team member",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		body := map[string]interface{}{}
		if v, _ := cmd.Flags().GetString("email"); cmd.Flags().Changed("email") {
			body["email"] = v
		}
		if v, _ := cmd.Flags().GetString("first-name"); cmd.Flags().Changed("first-name") {
			body["first_name"] = v
		}
		if v, _ := cmd.Flags().GetString("last-name"); cmd.Flags().Changed("last-name") {
			body["last_name"] = v
		}
		if v, _ := cmd.Flags().GetString("time-zone"); cmd.Flags().Changed("time-zone") {
			body["time_zone"] = v
		}
		if v, _ := cmd.Flags().GetString("role"); cmd.Flags().Changed("role") {
			body["organization_role"] = v
		}

		if len(body) == 0 {
			return fmt.Errorf("Nothing to update. Pass at least one of --email, --first-name, --last-name, --time-zone or --role.")
		}

		data, err := c.Patch(fmt.Sprintf("/team-members/%s", args[0]), body)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var teamMembersDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Remove a team member",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := c.Delete(fmt.Sprintf("/team-members/%s", args[0])); err != nil {
			return err
		}

		printMessage("Team member removed.")
		return nil
	},
}

func init() {
	addPaginationFlags(teamMembersListCmd)
	teamMembersListCmd.Flags().String("email", "", "Filter by email address")

	teamMembersCreateCmd.Flags().StringArray("email", nil, "Email address to invite (repeatable)")
	teamMembersCreateCmd.Flags().String("role", "", "Organization role")
	teamMembersCreateCmd.Flags().Bool("send-invitation-email", true, "Send invitation email")
	teamMembersCreateCmd.Flags().String("invited-by", "", "Inviter name or email")
	_ = teamMembersCreateCmd.MarkFlagRequired("email")
	_ = teamMembersCreateCmd.MarkFlagRequired("role")

	teamMembersUpdateCmd.Flags().String("email", "", "New email address")
	teamMembersUpdateCmd.Flags().String("first-name", "", "First name")
	teamMembersUpdateCmd.Flags().String("last-name", "", "Last name")
	teamMembersUpdateCmd.Flags().String("time-zone", "", "Time zone")
	teamMembersUpdateCmd.Flags().String("role", "", "Organization role")

	teamMembersCmd.AddCommand(teamMembersListCmd)
	teamMembersCmd.AddCommand(teamMembersShowCmd)
	teamMembersCmd.AddCommand(teamMembersCreateCmd)
	teamMembersCmd.AddCommand(teamMembersUpdateCmd)
	teamMembersCmd.AddCommand(teamMembersDeleteCmd)
	register(func(root *cobra.Command) { root.AddCommand(teamMembersCmd) })
}
