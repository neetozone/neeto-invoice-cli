package commands

import (
	"fmt"
	"net/url"

	"github.com/neetozone/neeto-invoice-cli/internal/output"
	"github.com/spf13/cobra"
)

var clientsCmd = &cobra.Command{
	Use:   "clients",
	Short: "Manage clients",
}

var clientsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List clients (find a client's id by name)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		if status, _ := cmd.Flags().GetString("status"); status != "" {
			params.Set("status", status)
		}
		if name, _ := cmd.Flags().GetString("name"); name != "" {
			params.Set("name", name)
		}

		data, err := c.Get("/clients", params)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "Show", Command: "neetoinvoice clients show <client-id>"},
		})
		return nil
	},
}

var clientsShowCmd = &cobra.Command{
	Use:   "show <client-id>",
	Short: "Show a client and its recipients",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/clients/%s", args[0]), nil)
		if err != nil {
			return err
		}

		printResource(data, []output.Breadcrumb{
			{Label: "Update", Command: "neetoinvoice clients update <client-id> --name <name>"},
			{Label: "Add recipient", Command: "neetoinvoice recipients create --client <client-id> --name <name> --email <email> --user-email <email>"},
		})
		return nil
	},
}

var clientsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a client",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Post("/clients", map[string]interface{}{"client": clientBody(cmd)})
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{
			{Label: "Show", Command: "neetoinvoice clients show <client-id>"},
		})
		return nil
	},
}

var clientsUpdateCmd = &cobra.Command{
	Use:   "update <client-id>",
	Short: "Update a client",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		data, err := c.Patch(fmt.Sprintf("/clients/%s", args[0]), map[string]interface{}{"client": clientBody(cmd)})
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{
			{Label: "Show", Command: "neetoinvoice clients show <client-id>"},
		})
		return nil
	},
}

func clientBody(cmd *cobra.Command) map[string]interface{} {
	body := map[string]interface{}{}

	if cmd.Flags().Changed("name") {
		body["name"], _ = cmd.Flags().GetString("name")
	}
	if cmd.Flags().Changed("secondary-name") {
		body["secondary_name"], _ = cmd.Flags().GetString("secondary-name")
	}
	if cmd.Flags().Changed("internal-notes") {
		body["internal_notes"], _ = cmd.Flags().GetString("internal-notes")
	}
	if cmd.Flags().Changed("status") {
		body["status"], _ = cmd.Flags().GetString("status")
	}
	if cmd.Flags().Changed("due-in-days") {
		body["default_invoice_due_in_days"], _ = cmd.Flags().GetInt("due-in-days")
	}
	if cmd.Flags().Changed("address") {
		address, _ := cmd.Flags().GetString("address")
		body["address_attributes"] = map[string]interface{}{"full_address": address}
	}

	return body
}

func addClientFlags(cmd *cobra.Command) {
	cmd.Flags().String("name", "", "Client name")
	cmd.Flags().String("secondary-name", "", "Secondary name")
	cmd.Flags().String("internal-notes", "", "Internal notes")
	cmd.Flags().String("status", "", "Status (active/archived)")
	cmd.Flags().Int("due-in-days", 0, "Default invoice due in days")
	cmd.Flags().String("address", "", "Full address")
}

func init() {
	addClientFlags(clientsCreateCmd)
	_ = clientsCreateCmd.MarkFlagRequired("name")

	addClientFlags(clientsUpdateCmd)

	clientsListCmd.Flags().String("status", "", "Filter by status: active or archived")
	clientsListCmd.Flags().String("name", "", "Filter by name (substring match)")

	clientsCmd.AddCommand(clientsListCmd)
	clientsCmd.AddCommand(clientsShowCmd)
	clientsCmd.AddCommand(clientsCreateCmd)
	clientsCmd.AddCommand(clientsUpdateCmd)
	rootCmd.AddCommand(clientsCmd)
}
