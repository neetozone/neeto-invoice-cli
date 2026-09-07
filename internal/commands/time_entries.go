package commands

import (
	"github.com/neetozone/neeto-cli-commons/output"
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
		if userEmail, _ := cmd.Flags().GetString("user-email"); userEmail != "" {
			params.Set("email", userEmail)
		}

		data, err := c.Get("/time-entries", params)
		if err != nil {
			return err
		}

		printListWithInlinePagination(data, "time_entries", inlinePaginationKeys{
			TotalRecords: "total_count",
			TotalPages:   "total_pages",
			CurrentPage:  "page",
			PageSize:     "page_size",
		}, []output.Breadcrumb{
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

		data, err := c.Post("/time-entries", timeEntryCreateBody(cmd))
		if err != nil {
			return err
		}

		printCreateResult(data, []output.Breadcrumb{
			{Label: "List", Command: "neetoinvoice time-entries list --client <client-id> --project <project-id>"},
		})
		return nil
	},
}

var timeEntriesUpdateCmd = &cobra.Command{
	Use:   "update <time-entry-id>",
	Short: "Update a time entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Put("/time-entries/"+args[0], timeEntryUpdateBody(cmd))
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{
			{Label: "List", Command: "neetoinvoice time-entries list --client <client-id> --project <project-id>"},
		})
		return nil
	},
}

var timeEntriesDeleteCmd = &cobra.Command{
	Use:   "delete <time-entry-id>",
	Short: "Delete a time entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		if err := c.Delete("/time-entries/" + args[0]); err != nil {
			return err
		}

		printMessage("Time entry deleted.")
		return nil
	},
}

func init() {
	addPaginationFlags(timeEntriesListCmd)
	timeEntriesListCmd.Flags().String("client", "", "Client identifier")
	timeEntriesListCmd.Flags().String("project", "", "Project identifier")
	timeEntriesListCmd.Flags().String("start-date", "", "Filter entries recorded on or after this date (YYYY-MM-DD)")
	timeEntriesListCmd.Flags().String("end-date", "", "Filter entries recorded on or before this date (YYYY-MM-DD)")
	timeEntriesListCmd.Flags().String("user-email", "", "Filter to a single user's entries by email")
	_ = timeEntriesListCmd.MarkFlagRequired("client")
	_ = timeEntriesListCmd.MarkFlagRequired("project")

	timeEntriesCreateCmd.Flags().String("client", "", "Client identifier")
	timeEntriesCreateCmd.Flags().String("project", "", "Project identifier")
	timeEntriesCreateCmd.Flags().String("task-id", "", "Task ID")
	timeEntriesCreateCmd.Flags().String("user-email", "", "Email of the organization user logging the time")
	timeEntriesCreateCmd.Flags().String("recorded-on", "", "Date the time was recorded (YYYY-MM-DD)")
	timeEntriesCreateCmd.Flags().Float64("hours", 0, "Hours spent (omit for no-hours projects)")
	timeEntriesCreateCmd.Flags().String("notes", "", "Notes")
	timeEntriesCreateCmd.Flags().Bool("is-override", false, "Override autolock for a locked date")
	_ = timeEntriesCreateCmd.MarkFlagRequired("client")
	_ = timeEntriesCreateCmd.MarkFlagRequired("project")
	_ = timeEntriesCreateCmd.MarkFlagRequired("task-id")
	_ = timeEntriesCreateCmd.MarkFlagRequired("user-email")
	_ = timeEntriesCreateCmd.MarkFlagRequired("recorded-on")

	timeEntriesUpdateCmd.Flags().String("notes", "", "New notes")
	timeEntriesUpdateCmd.Flags().Float64("hours", 0, "New hours")
	timeEntriesUpdateCmd.Flags().String("recorded-on", "", "New date (YYYY-MM-DD)")
	timeEntriesUpdateCmd.Flags().Bool("is-override", false, "Override autolock for a locked date")
	timeEntriesUpdateCmd.Flags().String("user-email", "", "Acting user's email")

	timeEntriesCmd.AddCommand(timeEntriesListCmd)
	timeEntriesCmd.AddCommand(timeEntriesCreateCmd)
	timeEntriesCmd.AddCommand(timeEntriesUpdateCmd)
	timeEntriesCmd.AddCommand(timeEntriesDeleteCmd)
	register(func(root *cobra.Command) { root.AddCommand(timeEntriesCmd) })
}

func timeEntryCreateBody(cmd *cobra.Command) map[string]interface{} {
	clientID, _ := cmd.Flags().GetString("client")
	projectID, _ := cmd.Flags().GetString("project")
	taskID, _ := cmd.Flags().GetString("task-id")
	userEmail, _ := cmd.Flags().GetString("user-email")
	recordedOn, _ := cmd.Flags().GetString("recorded-on")

	body := map[string]interface{}{
		"client_id":   clientID,
		"project_id":  projectID,
		"task_id":     taskID,
		"email":       userEmail,
		"recorded_on": recordedOn,
	}

	if cmd.Flags().Changed("hours") {
		hours, _ := cmd.Flags().GetFloat64("hours")
		body["hours"] = hours
	}
	if cmd.Flags().Changed("is-override") {
		body["is_override"] = true
	}
	if notes, _ := cmd.Flags().GetString("notes"); notes != "" {
		body["notes"] = notes
	}

	return body
}

func timeEntryUpdateBody(cmd *cobra.Command) map[string]interface{} {
	body := map[string]interface{}{}

	if cmd.Flags().Changed("notes") {
		notes, _ := cmd.Flags().GetString("notes")
		body["notes"] = notes
	}
	if cmd.Flags().Changed("hours") {
		hours, _ := cmd.Flags().GetFloat64("hours")
		body["hours"] = hours
	}
	if recordedOn, _ := cmd.Flags().GetString("recorded-on"); recordedOn != "" {
		body["recorded_on"] = recordedOn
	}
	if cmd.Flags().Changed("is-override") {
		body["is_override"] = true
	}
	if userEmail, _ := cmd.Flags().GetString("user-email"); userEmail != "" {
		body["email"] = userEmail
	}

	return body
}
