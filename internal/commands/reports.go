package commands

import (
	"net/url"
	"strconv"

	"github.com/spf13/cobra"
)

var reportsCmd = &cobra.Command{
	Use:   "reports",
	Short: "Timesheet, payroll and PTO reports",
}

var reportsPayrollSummaryCmd = &cobra.Command{
	Use:   "payroll-summary",
	Short: "Per-user working days and LOP days for a month",
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

		data, err := c.Get("/reports/payroll-summary", params)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var reportsMissingEntriesCmd = &cobra.Command{
	Use:   "missing-entries",
	Short: "Working days with no timesheet entry for a user",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		email, _ := cmd.Flags().GetString("user-email")
		startDate, _ := cmd.Flags().GetString("start-date")
		endDate, _ := cmd.Flags().GetString("end-date")
		params.Set("email", email)
		params.Set("start_date", startDate)
		params.Set("end_date", endDate)

		data, err := c.Get("/reports/missing-entries", params)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

var reportsTimesheetSummaryCmd = &cobra.Command{
	Use:   "timesheet-summary",
	Short: "Per-user, per-day, per-project hours for a date range",
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

		data, err := c.Get("/reports/timesheet-summary", params)
		if err != nil {
			return err
		}

		printResource(data, nil)
		return nil
	},
}

func init() {
	reportsPayrollSummaryCmd.Flags().Int("month", 0, "Month number (1-12)")
	reportsPayrollSummaryCmd.Flags().Int("year", 0, "Year, e.g. 2026")
	_ = reportsPayrollSummaryCmd.MarkFlagRequired("month")
	_ = reportsPayrollSummaryCmd.MarkFlagRequired("year")

	reportsMissingEntriesCmd.Flags().String("user-email", "", "User email")
	reportsMissingEntriesCmd.Flags().String("start-date", "", "Range start (YYYY-MM-DD)")
	reportsMissingEntriesCmd.Flags().String("end-date", "", "Range end (YYYY-MM-DD)")
	_ = reportsMissingEntriesCmd.MarkFlagRequired("user-email")
	_ = reportsMissingEntriesCmd.MarkFlagRequired("start-date")
	_ = reportsMissingEntriesCmd.MarkFlagRequired("end-date")

	reportsTimesheetSummaryCmd.Flags().String("start-date", "", "Range start (YYYY-MM-DD)")
	reportsTimesheetSummaryCmd.Flags().String("end-date", "", "Range end (YYYY-MM-DD)")
	reportsTimesheetSummaryCmd.Flags().String("user-email", "", "Filter to a single user by email")
	_ = reportsTimesheetSummaryCmd.MarkFlagRequired("start-date")
	_ = reportsTimesheetSummaryCmd.MarkFlagRequired("end-date")

	reportsCmd.AddCommand(reportsPayrollSummaryCmd)
	reportsCmd.AddCommand(reportsMissingEntriesCmd)
	reportsCmd.AddCommand(reportsTimesheetSummaryCmd)
	rootCmd.AddCommand(reportsCmd)
}
