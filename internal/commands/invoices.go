package commands

import (
	"fmt"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var invoicesCmd = &cobra.Command{
	Use:   "invoices",
	Short: "Manage invoices",
}

var invoicesCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Generate an invoice for a client",
	Long: `Generate an invoice for a client from its unbilled time entries, services, and expenses.

Scalar fields are available as flags. For line items (invoice_time_entries,
invoice_services, invoice_expenses, tax_details, custom_attributes, invoice_details)
pass a JSON file via --json-file; flag values override the file's top-level keys.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		clientID, _ := cmd.Flags().GetString("client")

		body := map[string]interface{}{}
		dataFile, _ := cmd.Flags().GetString("json-file")
		if dataFile == "" {
			dataFile, _ = cmd.Flags().GetString("data")
		}
		if dataFile != "" {
			body, err = readJSONFile(dataFile)
			if err != nil {
				return err
			}
		}

		userEmail, _ := cmd.Flags().GetString("user-email")
		body["email"] = userEmail

		if cmd.Flags().Changed("issue-date") {
			body["issue_date"], _ = cmd.Flags().GetString("issue-date")
		}
		if cmd.Flags().Changed("due-date") {
			body["due_date"], _ = cmd.Flags().GetString("due-date")
		}
		if cmd.Flags().Changed("number") {
			body["number"], _ = cmd.Flags().GetString("number")
		}
		if cmd.Flags().Changed("notes") {
			body["notes"], _ = cmd.Flags().GetString("notes")
		}
		if cmd.Flags().Changed("project-id") {
			body["project_id"], _ = cmd.Flags().GetStringSlice("project-id")
		}
		if cmd.Flags().Changed("send-email") {
			body["send_email"], _ = cmd.Flags().GetBool("send-email")
		}

		data, err := c.Post(fmt.Sprintf("/clients/%s/invoices", clientID), body)
		if err != nil {
			return err
		}

		printCreateResult(data, []output.Breadcrumb{
			{Label: "Client", Command: "neetoinvoice clients show <client-id>"},
		})
		return nil
	},
}

func init() {
	invoicesCreateCmd.Flags().String("client", "", "Client identifier")
	invoicesCreateCmd.Flags().String("user-email", "", "Email of the acting organization user")
	invoicesCreateCmd.Flags().String("issue-date", "", "Issue date (MM/DD/YYYY)")
	invoicesCreateCmd.Flags().String("due-date", "", "Due date (MM/DD/YYYY)")
	invoicesCreateCmd.Flags().String("number", "", "Invoice number")
	invoicesCreateCmd.Flags().String("notes", "", "Invoice notes")
	invoicesCreateCmd.Flags().StringSlice("project-id", nil, "Project ID(s) to invoice (repeatable)")
	invoicesCreateCmd.Flags().Bool("send-email", false, "Email the invoice to the client's recipients")
	invoicesCreateCmd.Flags().String("json-file", "", "Path to a JSON file with the full invoice payload (line items, taxes, email details)")
	invoicesCreateCmd.Flags().String("data", "", "Path to a JSON file with the full invoice payload (line items, taxes, email details)")
	_ = invoicesCreateCmd.Flags().MarkDeprecated("data", "use --json-file instead")
	_ = invoicesCreateCmd.MarkFlagRequired("client")
	_ = invoicesCreateCmd.MarkFlagRequired("user-email")

	invoicesCmd.AddCommand(invoicesCreateCmd)
	register(func(root *cobra.Command) { root.AddCommand(invoicesCmd) })
}
