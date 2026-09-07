package commands

import (
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

var monthlyPtosCmd = &cobra.Command{
	Use:   "monthly-ptos",
	Short: "View and manage the monthly PTO report",
}

var monthlyPtosListCmd = &cobra.Command{
	Use:   "list",
	Short: "List the monthly PTO report for a month and year",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		month, _ := cmd.Flags().GetInt("month")
		year, _ := cmd.Flags().GetInt("year")
		params.Set("month", strconv.Itoa(month))
		params.Set("year", strconv.Itoa(year))
		if email, _ := cmd.Flags().GetString("user-email"); email != "" {
			params.Set("email", email)
		}

		data, err := c.Get("/monthly-ptos", params)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var monthlyPtosUpdateEarnedCmd = &cobra.Command{
	Use:   "update-earned",
	Short: "Set PTO earned for users (e.g. 0 to remove earned PTO)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		month, _ := cmd.Flags().GetInt("month")
		year, _ := cmd.Flags().GetInt("year")
		emails, _ := cmd.Flags().GetStringSlice("email")

		body := map[string]interface{}{
			"emails": emails,
			"month":  month,
			"year":   year,
		}
		if cmd.Flags().Changed("pto-earned") {
			ptoEarned, _ := cmd.Flags().GetFloat64("pto-earned")
			body["pto_earned"] = ptoEarned
		}

		data, err := c.Patch("/monthly-ptos/update-earned", body)
		if err != nil {
			return err
		}

		printActionResult(data, nil)
		return nil
	},
}

func init() {
	monthlyPtosListCmd.Flags().Int("month", 0, "Month number (1-12)")
	monthlyPtosListCmd.Flags().Int("year", 0, "Year, e.g. 2026")
	monthlyPtosListCmd.Flags().String("user-email", "", "Filter to a single user by email")
	_ = monthlyPtosListCmd.MarkFlagRequired("month")
	_ = monthlyPtosListCmd.MarkFlagRequired("year")

	monthlyPtosUpdateEarnedCmd.Flags().StringSlice("email", []string{}, "User email (repeatable)")
	monthlyPtosUpdateEarnedCmd.Flags().Int("month", 0, "Month number (1-12)")
	monthlyPtosUpdateEarnedCmd.Flags().Int("year", 0, "Year, e.g. 2026")
	monthlyPtosUpdateEarnedCmd.Flags().Float64("pto-earned", 0, "New PTO earned in hours (default: 0)")
	_ = monthlyPtosUpdateEarnedCmd.MarkFlagRequired("email")
	_ = monthlyPtosUpdateEarnedCmd.MarkFlagRequired("month")
	_ = monthlyPtosUpdateEarnedCmd.MarkFlagRequired("year")

	monthlyPtosCmd.AddCommand(monthlyPtosListCmd)
	monthlyPtosCmd.AddCommand(monthlyPtosUpdateEarnedCmd)
	register(func(root *cobra.Command) { root.AddCommand(monthlyPtosCmd) })
}
