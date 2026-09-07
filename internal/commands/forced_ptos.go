package commands

import (
	"net/url"

	"github.com/spf13/cobra"
)

var forcedPtosCmd = &cobra.Command{
	Use:   "forced-ptos",
	Short: "View and create Forced PTO entries",
}

var forcedPtosListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Forced-PTO entries in a date range",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")
		params.Set("start_date", startDate)
		params.Set("end_date", endDate)
		if email, _ := cmd.Flags().GetString("user-email"); email != "" {
			params.Set("email", email)
		}

		data, err := c.Get("/forced-ptos", params)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var forcedPtosCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Mark a day as Forced PTO for a user",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		email, _ := cmd.Flags().GetString("user-email")
		date, _ := cmd.Flags().GetString("date")

		body := map[string]interface{}{
			"email": email,
			"date":  date,
		}
		if cmd.Flags().Changed("hours") {
			hours, _ := cmd.Flags().GetFloat64("hours")
			body["hours"] = hours
		}

		data, err := c.Post("/forced-ptos", body)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

func init() {
	forcedPtosListCmd.Flags().String("start-date", "", "Range start (YYYY-MM-DD)")
	forcedPtosListCmd.Flags().String("end-date", "", "Range end (YYYY-MM-DD)")
	forcedPtosListCmd.Flags().String("user-email", "", "Filter to a single user by email")
	_ = forcedPtosListCmd.MarkFlagRequired("start-date")
	_ = forcedPtosListCmd.MarkFlagRequired("end-date")

	forcedPtosCreateCmd.Flags().String("user-email", "", "User email")
	forcedPtosCreateCmd.Flags().String("date", "", "Date to mark as Forced PTO (YYYY-MM-DD)")
	forcedPtosCreateCmd.Flags().Float64("hours", 0, "Hours to log (default: 8)")
	_ = forcedPtosCreateCmd.MarkFlagRequired("user-email")
	_ = forcedPtosCreateCmd.MarkFlagRequired("date")

	forcedPtosCmd.AddCommand(forcedPtosListCmd)
	forcedPtosCmd.AddCommand(forcedPtosCreateCmd)
	register(func(root *cobra.Command) { root.AddCommand(forcedPtosCmd) })
}
