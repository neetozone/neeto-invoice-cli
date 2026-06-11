package commands

import (
	"fmt"

	"github.com/neetozone/neeto-invoice-cli/internal/output"
	"github.com/spf13/cobra"
)

var projectUsersCmd = &cobra.Command{
	Use:   "project-users",
	Short: "Manage project users",
}

var projectUsersListCmd = &cobra.Command{
	Use:   "list",
	Short: "List users on a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		projectID, _ := cmd.Flags().GetString("project")

		data, err := c.Get(fmt.Sprintf("/projects/%s/project-users", projectID), nil)
		if err != nil {
			return err
		}

		printList(data, "project_users", []output.Breadcrumb{
			{Label: "Add user", Command: "neetoinvoice project-users create --project <project-id> --user-id <user-id>"},
			{Label: "Update", Command: "neetoinvoice project-users update <project-user-id> --project <project-id>"},
		})
		return nil
	},
}

var projectUsersCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Add a user to a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		projectID, _ := cmd.Flags().GetString("project")

		data, err := c.Post(
			fmt.Sprintf("/projects/%s/project-users", projectID),
			map[string]interface{}{"project_user": projectUserBody(cmd)},
		)
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{
			{Label: "List", Command: "neetoinvoice project-users list --project <project-id>"},
		})
		return nil
	},
}

var projectUsersUpdateCmd = &cobra.Command{
	Use:   "update <project-user-id>",
	Short: "Update a project user",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		projectID, _ := cmd.Flags().GetString("project")

		data, err := c.Patch(
			fmt.Sprintf("/projects/%s/project-users/%s", projectID, args[0]),
			map[string]interface{}{"project_user": projectUserBody(cmd)},
		)
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{
			{Label: "List", Command: "neetoinvoice project-users list --project <project-id>"},
		})
		return nil
	},
}

var projectUsersDeleteCmd = &cobra.Command{
	Use:   "delete <project-user-id>",
	Short: "Remove a user from a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		projectID, _ := cmd.Flags().GetString("project")

		if err := c.Delete(fmt.Sprintf("/projects/%s/project-users/%s", projectID, args[0])); err != nil {
			return err
		}

		output.PrintMessage("Project user removed.")
		return nil
	},
}

func projectUserBody(cmd *cobra.Command) map[string]interface{} {
	body := map[string]interface{}{}

	if cmd.Flags().Changed("user-id") {
		body["user_id"], _ = cmd.Flags().GetString("user-id")
	}
	if cmd.Flags().Changed("role") {
		body["role"], _ = cmd.Flags().GetString("role")
	}
	if cmd.Flags().Changed("hourly-rate") {
		body["hourly_rate"], _ = cmd.Flags().GetFloat64("hourly-rate")
	}

	return body
}

func init() {
	projectUsersListCmd.Flags().String("project", "", "Project identifier")
	_ = projectUsersListCmd.MarkFlagRequired("project")

	projectUsersCreateCmd.Flags().String("project", "", "Project identifier")
	projectUsersCreateCmd.Flags().String("user-id", "", "User ID to add")
	projectUsersCreateCmd.Flags().String("role", "", "Role (regular_user/project_manager)")
	projectUsersCreateCmd.Flags().Float64("hourly-rate", 0, "Hourly rate")
	_ = projectUsersCreateCmd.MarkFlagRequired("project")
	_ = projectUsersCreateCmd.MarkFlagRequired("user-id")

	projectUsersUpdateCmd.Flags().String("project", "", "Project identifier")
	projectUsersUpdateCmd.Flags().String("role", "", "Role (regular_user/project_manager)")
	projectUsersUpdateCmd.Flags().Float64("hourly-rate", 0, "Hourly rate")
	_ = projectUsersUpdateCmd.MarkFlagRequired("project")

	projectUsersDeleteCmd.Flags().String("project", "", "Project identifier")
	_ = projectUsersDeleteCmd.MarkFlagRequired("project")

	projectUsersCmd.AddCommand(projectUsersListCmd)
	projectUsersCmd.AddCommand(projectUsersCreateCmd)
	projectUsersCmd.AddCommand(projectUsersUpdateCmd)
	projectUsersCmd.AddCommand(projectUsersDeleteCmd)
	rootCmd.AddCommand(projectUsersCmd)
}
