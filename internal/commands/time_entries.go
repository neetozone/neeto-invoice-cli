package commands

import (
	"github.com/neetozone/neeto-invoice-cli/internal/output"
	"github.com/spf13/cobra"
)

var timeEntriesCmd = &cobra.Command{
	Use:   "time-entries",
	Short: "Manage time entries",
}

var timeEntriesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List unbilled time entries for a project",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := paginationParams(cmd)
		clientID, _ := cmd.Flags().GetString("client")
		projectID, _ := cmd.Flags().GetString("project")
		params.Set("client_id", clientID)
		params.Set("project_id", projectID)

		if startDate, _ := cmd.Flags().GetString("start-date"); startDate != "" {
			params.Set("start_date", startDate)
		}
		if endDate, _ := cmd.Flags().GetString("end-date"); endDate != "" {
			params.Set("end_date", endDate)
		}

		data, err := c.Get("/time-entries", params)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "Log time", Command: "neetoinvoice time-entries create --client <client-id> --project <project-id> --task-id <task-id> --user-email <email> --recorded-on <date> --hours <hours>"},
		})
		return nil
	},
}

var timeEntriesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Log a time entry",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		clientID, _ := cmd.Flags().GetString("client")
		projectID, _ := cmd.Flags().GetString("project")
		taskID, _ := cmd.Flags().GetString("task-id")
		userEmail, _ := cmd.Flags().GetString("user-email")
		recordedOn, _ := cmd.Flags().GetString("recorded-on")
		hours, _ := cmd.Flags().GetFloat64("hours")

		body := map[string]interface{}{
			"client_id":   clientID,
			"project_id":  projectID,
			"task_id":     taskID,
			"email":       userEmail,
			"recorded_on": recordedOn,
			"hours":       hours,
		}

		if notes, _ := cmd.Flags().GetString("notes"); notes != "" {
			body["notes"] = notes
		}

		data, err := c.Post("/time-entries", body)
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{
			{Label: "List", Command: "neetoinvoice time-entries list --client <client-id> --project <project-id>"},
		})
		return nil
	},
}

func init() {
	addPaginationFlags(timeEntriesListCmd)
	timeEntriesListCmd.Flags().String("client", "", "Client identifier")
	timeEntriesListCmd.Flags().String("project", "", "Project identifier")
	timeEntriesListCmd.Flags().String("start-date", "", "Filter entries recorded on or after this date (YYYY-MM-DD)")
	timeEntriesListCmd.Flags().String("end-date", "", "Filter entries recorded on or before this date (YYYY-MM-DD)")
	_ = timeEntriesListCmd.MarkFlagRequired("client")
	_ = timeEntriesListCmd.MarkFlagRequired("project")

	timeEntriesCreateCmd.Flags().String("client", "", "Client identifier")
	timeEntriesCreateCmd.Flags().String("project", "", "Project identifier")
	timeEntriesCreateCmd.Flags().String("task-id", "", "Task ID")
	timeEntriesCreateCmd.Flags().String("user-email", "", "Email of the organization user logging the time")
	timeEntriesCreateCmd.Flags().String("recorded-on", "", "Date the time was recorded (YYYY-MM-DD)")
	timeEntriesCreateCmd.Flags().Float64("hours", 0, "Hours spent")
	timeEntriesCreateCmd.Flags().String("notes", "", "Notes")
	_ = timeEntriesCreateCmd.MarkFlagRequired("client")
	_ = timeEntriesCreateCmd.MarkFlagRequired("project")
	_ = timeEntriesCreateCmd.MarkFlagRequired("task-id")
	_ = timeEntriesCreateCmd.MarkFlagRequired("user-email")
	_ = timeEntriesCreateCmd.MarkFlagRequired("recorded-on")
	_ = timeEntriesCreateCmd.MarkFlagRequired("hours")

	timeEntriesCmd.AddCommand(timeEntriesListCmd)
	timeEntriesCmd.AddCommand(timeEntriesCreateCmd)
	rootCmd.AddCommand(timeEntriesCmd)
}
