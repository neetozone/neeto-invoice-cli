package commands

import (
	"fmt"

	"github.com/neetozone/neeto-cli-commons/output"
	"github.com/spf13/cobra"
)

var recipientsCmd = &cobra.Command{
	Use:   "recipients",
	Short: "Manage client recipients",
}

var recipientsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Add a recipient to a client",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		clientID, _ := cmd.Flags().GetString("client")
		userEmail, _ := cmd.Flags().GetString("user-email")

		body := map[string]interface{}{
			"recipient": recipientBody(cmd),
			"email":     userEmail,
		}

		data, err := c.Post(fmt.Sprintf("/clients/%s/recipients", clientID), body)
		if err != nil {
			return err
		}

		printCreateResult(data, []output.Breadcrumb{
			{Label: "Client", Command: "neetoinvoice clients show <client-id>"},
		})
		return nil
	},
}

var recipientsUpdateCmd = &cobra.Command{
	Use:   "update <recipient-id>",
	Short: "Update a client recipient",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		clientID, _ := cmd.Flags().GetString("client")

		data, err := c.Patch(
			fmt.Sprintf("/clients/%s/recipients/%s", clientID, args[0]),
			map[string]interface{}{"recipient": recipientBody(cmd)},
		)
		if err != nil {
			return err
		}

		printActionResult(data, []output.Breadcrumb{
			{Label: "Client", Command: "neetoinvoice clients show <client-id>"},
		})
		return nil
	},
}

var recipientsDeleteCmd = &cobra.Command{
	Use:   "delete <recipient-id>",
	Short: "Delete a client recipient",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := getClient(cmd)
		if err != nil {
			return err
		}

		clientID, _ := cmd.Flags().GetString("client")

		if err := c.Delete(fmt.Sprintf("/clients/%s/recipients/%s", clientID, args[0])); err != nil {
			return err
		}

		printMessage("Recipient deleted.")
		return nil
	},
}

func recipientBody(cmd *cobra.Command) map[string]interface{} {
	body := map[string]interface{}{}

	if cmd.Flags().Changed("name") {
		body["name"], _ = cmd.Flags().GetString("name")
	}
	if cmd.Flags().Changed("email") {
		body["email"], _ = cmd.Flags().GetString("email")
	}

	return body
}

func init() {
	recipientsCreateCmd.Flags().String("client", "", "Client identifier")
	recipientsCreateCmd.Flags().String("name", "", "Recipient name")
	recipientsCreateCmd.Flags().String("email", "", "Recipient email")
	recipientsCreateCmd.Flags().String("user-email", "", "Email of the acting organization user")
	_ = recipientsCreateCmd.MarkFlagRequired("client")
	_ = recipientsCreateCmd.MarkFlagRequired("name")
	_ = recipientsCreateCmd.MarkFlagRequired("email")
	_ = recipientsCreateCmd.MarkFlagRequired("user-email")

	recipientsUpdateCmd.Flags().String("client", "", "Client identifier")
	recipientsUpdateCmd.Flags().String("name", "", "Recipient name")
	recipientsUpdateCmd.Flags().String("email", "", "Recipient email")
	_ = recipientsUpdateCmd.MarkFlagRequired("client")

	recipientsDeleteCmd.Flags().String("client", "", "Client identifier")
	_ = recipientsDeleteCmd.MarkFlagRequired("client")

	recipientsCmd.AddCommand(recipientsCreateCmd)
	recipientsCmd.AddCommand(recipientsUpdateCmd)
	recipientsCmd.AddCommand(recipientsDeleteCmd)
	register(func(root *cobra.Command) { root.AddCommand(recipientsCmd) })
}
