package commands

import (
	"fmt"
	"net/url"

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
pass a JSON file via --json-file; flag values override the file's top-level keys.
invoice_expenses rows take a category name, and tax_details is an array of
name and percentage pairs. Dates are YYYY-MM-DD. With --send-email and no
invoice_details.recipient_ids, the invoice is emailed to every current recipient
of the client.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		clientID, _ := cmd.Flags().GetString("client")

		body := map[string]interface{}{}
		if payloadFile := invoicePayloadFile(cmd); payloadFile != "" {
			body, err = readJSONFile(payloadFile)
			if err != nil {
				return err
			}
		}

		if userEmail, _ := cmd.Flags().GetString("user-email"); userEmail != "" {
			body["email"] = userEmail
		}

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
		if cmd.Flags().Changed("time-entry-id") {
			body["time_entry_ids"], _ = cmd.Flags().GetStringSlice("time-entry-id")
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

var invoicesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List invoices, newest first",
	Long: `List invoices, newest issue date first. Filter by client, status, issue dates
or --overdue. Each row includes grand_total, amount_paid, amount_due and days_overdue.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get("/invoices", invoicesListParams(cmd))
		if err != nil {
			return err
		}

		printList(data, "invoices", []output.Breadcrumb{
			{Label: "Send", Command: "neetoinvoice invoices send <invoice> --client <client-id>"},
		})
		return nil
	},
}

var invoicesSendCmd = &cobra.Command{
	Use:   "send <invoice>",
	Short: "Email an invoice to the client's recipients",
	Long: `Email a generated invoice to the client's recipients. The invoice is the invoice
number, identifier or ID. Every current recipient is emailed unless --recipient-id
narrows it.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		clientID, _ := cmd.Flags().GetString("client")

		data, err := c.Post(
			fmt.Sprintf("/clients/%s/invoices/%s/send", url.PathEscape(clientID), url.PathEscape(args[0])),
			invoiceSendBody(cmd),
		)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "List", Command: "neetoinvoice invoices list --client <client-id>"},
		})
		return nil
	},
}

func init() {
	addPaginationFlags(invoicesListCmd)
	invoicesListCmd.Flags().String("client", "", "Only this client's invoices (identifier or ID)")
	invoicesListCmd.Flags().String("status", "", "Filter by status: generating, generated, paid, written_off, failed or pending_payment")
	invoicesListCmd.Flags().Bool("overdue", false, "Only unpaid invoices past their due date")
	invoicesListCmd.Flags().String("start-date", "", "Issued on or after this date (YYYY-MM-DD)")
	invoicesListCmd.Flags().String("end-date", "", "Issued on or before this date (YYYY-MM-DD)")

	invoicesSendCmd.Flags().String("client", "", "Client identifier or ID")
	invoicesSendCmd.Flags().StringSlice("recipient-id", nil, "Recipient ID to email (repeatable; default: every current recipient)")
	invoicesSendCmd.Flags().String("subject", "", "Email subject")
	invoicesSendCmd.Flags().String("message", "", "Email message")
	invoicesSendCmd.Flags().Bool("attach-pdf", false, "Attach the invoice PDF")
	invoicesSendCmd.Flags().Bool("send-me-a-copy", false, "Also email a copy to the acting user")
	invoicesSendCmd.Flags().String("user-email", "", "Email of the acting organization user (needed only with an API key)")
	_ = invoicesSendCmd.MarkFlagRequired("client")

	invoicesCreateCmd.Flags().String("client", "", "Client identifier")
	invoicesCreateCmd.Flags().String("user-email", "", "Email of the acting organization user (needed only with an API key)")
	invoicesCreateCmd.Flags().String("issue-date", "", "Issue date (YYYY-MM-DD)")
	invoicesCreateCmd.Flags().String("due-date", "", "Due date (YYYY-MM-DD)")
	invoicesCreateCmd.Flags().String("number", "", "Invoice number")
	invoicesCreateCmd.Flags().String("notes", "", "Invoice notes")
	invoicesCreateCmd.Flags().StringSlice("project-id", nil, "Project ID(s) to invoice (repeatable)")
	invoicesCreateCmd.Flags().StringSlice("time-entry-id", nil, "Unbilled time entry ID to bill (repeatable)")
	invoicesCreateCmd.Flags().Bool("send-email", false, "Email the invoice to the client's recipients")
	invoicesCreateCmd.Flags().String("json-file", "", "Path to a JSON file with the full invoice payload (line items, taxes, email details)")
	invoicesCreateCmd.Flags().String("data", "", "Path to a JSON file with the full invoice payload (line items, taxes, email details)")
	_ = invoicesCreateCmd.Flags().MarkDeprecated("data", "use --json-file instead")
	_ = invoicesCreateCmd.MarkFlagRequired("client")
	_ = invoicesCreateCmd.MarkFlagRequired("number")
	invoicesCreateCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		payloadFile := invoicePayloadFile(cmd)
		if payloadFile == "" {
			return nil
		}

		return cmd.Flags().Set("json-file", payloadFile)
	}
	allowJSONFileToSatisfyRequiredFlags(invoicesCreateCmd)

	invoicesCmd.AddCommand(invoicesListCmd)
	invoicesCmd.AddCommand(invoicesSendCmd)
	invoicesCmd.AddCommand(invoicesCreateCmd)
	register(func(root *cobra.Command) { root.AddCommand(invoicesCmd) })
}

func invoicePayloadFile(cmd *cobra.Command) string {
	if jsonFile, _ := cmd.Flags().GetString("json-file"); jsonFile != "" {
		return jsonFile
	}

	dataFile, _ := cmd.Flags().GetString("data")
	return dataFile
}

func invoicesListParams(cmd *cobra.Command) url.Values {
	params := paginationParams(cmd)

	for flag, key := range map[string]string{
		"client":     "client_id",
		"status":     "status",
		"start-date": "start_date",
		"end-date":   "end_date",
	} {
		if value, _ := cmd.Flags().GetString(flag); value != "" {
			params.Set(key, value)
		}
	}
	if overdue, _ := cmd.Flags().GetBool("overdue"); overdue {
		params.Set("overdue", "true")
	}

	return params
}

func invoiceSendBody(cmd *cobra.Command) map[string]interface{} {
	body := map[string]interface{}{}

	if cmd.Flags().Changed("recipient-id") {
		body["recipient_ids"], _ = cmd.Flags().GetStringSlice("recipient-id")
	}
	if cmd.Flags().Changed("subject") {
		body["subject"], _ = cmd.Flags().GetString("subject")
	}
	if cmd.Flags().Changed("message") {
		body["message"], _ = cmd.Flags().GetString("message")
	}
	if cmd.Flags().Changed("attach-pdf") {
		body["attach_pdf_invoice"], _ = cmd.Flags().GetBool("attach-pdf")
	}
	if cmd.Flags().Changed("send-me-a-copy") {
		body["send_me_a_copy"], _ = cmd.Flags().GetBool("send-me-a-copy")
	}
	if userEmail, _ := cmd.Flags().GetString("user-email"); userEmail != "" {
		body["email"] = userEmail
	}

	return body
}
