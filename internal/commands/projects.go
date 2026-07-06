package commands

import (
	"fmt"
	"net/url"

	"github.com/neetozone/neeto-invoice-cli/internal/output"
	"github.com/spf13/cobra"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects",
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects (find a project's id, or a user's projects)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		if clientID, _ := cmd.Flags().GetString("client-id"); clientID != "" {
			params.Set("client_id", clientID)
		}
		if userEmail, _ := cmd.Flags().GetString("user-email"); userEmail != "" {
			params.Set("email", userEmail)
		}
		if status, _ := cmd.Flags().GetString("status"); status != "" {
			params.Set("status", status)
		}

		data, err := c.Get("/projects", params)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "Show", Command: "neetoinvoice projects show <project-id>"},
		})
		return nil
	},
}

var projectsShowCmd = &cobra.Command{
	Use:   "show <project-id>",
	Short: "Show a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/projects/%s", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "Project users", Command: "neetoinvoice project-users list --project <project-id>"},
			{Label: "Update", Command: "neetoinvoice projects update <project-id> --user-email <email>"},
		})
		return nil
	},
}

var projectsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		userEmail, _ := cmd.Flags().GetString("user-email")
		body := map[string]interface{}{
			"project": projectBody(cmd),
			"email":   userEmail,
		}

		data, err := c.Post("/projects", body)
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{
			{Label: "Show", Command: "neetoinvoice projects show <project-id>"},
			{Label: "Add user", Command: "neetoinvoice project-users create --project <project-id> --user-id <user-id>"},
		})
		return nil
	},
}

var projectsUpdateCmd = &cobra.Command{
	Use:   "update <project-id>",
	Short: "Update a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		userEmail, _ := cmd.Flags().GetString("user-email")
		body := map[string]interface{}{
			"project": projectBody(cmd),
			"email":   userEmail,
		}

		data, err := c.Patch(fmt.Sprintf("/projects/%s", args[0]), body)
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{
			{Label: "Show", Command: "neetoinvoice projects show <project-id>"},
		})
		return nil
	},
}

func projectBody(cmd *cobra.Command) map[string]interface{} {
	body := map[string]interface{}{}

	if cmd.Flags().Changed("name") {
		body["name"], _ = cmd.Flags().GetString("name")
	}
	if cmd.Flags().Changed("client-id") {
		body["client_id"], _ = cmd.Flags().GetString("client-id")
	}
	if cmd.Flags().Changed("billing-method") {
		body["billing_method"], _ = cmd.Flags().GetString("billing-method")
	}
	if cmd.Flags().Changed("hourly-rate") {
		body["hourly_rate"], _ = cmd.Flags().GetFloat64("hourly-rate")
	}
	if cmd.Flags().Changed("flat-amount") {
		body["flat_amount"], _ = cmd.Flags().GetFloat64("flat-amount")
	}
	if cmd.Flags().Changed("currency") {
		body["currency"], _ = cmd.Flags().GetString("currency")
	}
	if cmd.Flags().Changed("internal-notes") {
		body["internal_notes"], _ = cmd.Flags().GetString("internal-notes")
	}
	if cmd.Flags().Changed("task") {
		taskNames, _ := cmd.Flags().GetStringSlice("task")
		tasks := make([]map[string]interface{}, 0, len(taskNames))
		for _, name := range taskNames {
			tasks = append(tasks, map[string]interface{}{"name": name})
		}
		body["tasks_attributes"] = tasks
	}

	return body
}

func addProjectFlags(cmd *cobra.Command) {
	cmd.Flags().String("name", "", "Project name")
	cmd.Flags().String("client-id", "", "Client record ID (UUID) the project belongs to")
	cmd.Flags().String("billing-method", "", "Billing method (hourly_project_rate, hourly_person_rate, hourly_task_rate, fixed_price_project)")
	cmd.Flags().Float64("hourly-rate", 0, "Hourly rate")
	cmd.Flags().Float64("flat-amount", 0, "Flat amount (for fixed price projects)")
	cmd.Flags().String("currency", "", "Currency code (e.g. USD)")
	cmd.Flags().String("internal-notes", "", "Internal notes")
	cmd.Flags().StringSlice("task", nil, "Task name to add to the project (repeatable; at least one task is required on create)")
	cmd.Flags().String("user-email", "", "Email of the acting organization user")
	_ = cmd.MarkFlagRequired("user-email")
}

func init() {
	addProjectFlags(projectsCreateCmd)
	_ = projectsCreateCmd.MarkFlagRequired("name")
	_ = projectsCreateCmd.MarkFlagRequired("client-id")
	_ = projectsCreateCmd.MarkFlagRequired("task")

	addProjectFlags(projectsUpdateCmd)

	projectsListCmd.Flags().String("client-id", "", "Filter to one client (identifier or ID)")
	projectsListCmd.Flags().String("user-email", "", "Filter to a user's projects")
	projectsListCmd.Flags().String("status", "", "Filter by status: active or archived")

	projectsCmd.AddCommand(projectsListCmd)
	projectsCmd.AddCommand(projectsShowCmd)
	projectsCmd.AddCommand(projectsCreateCmd)
	projectsCmd.AddCommand(projectsUpdateCmd)
	rootCmd.AddCommand(projectsCmd)
}
