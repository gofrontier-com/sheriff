package plan

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofrontier-com/go-utils/output"
	"github.com/gofrontier-com/sheriff/pkg/cmd/app/apply"
	"github.com/spf13/cobra"
)

var (
	configDir      string
	subscriptionId string
)

// NewCmdPlanAzureRm creates a command to llan the Azure RM config changes
func NewCmdPlanAzureRm() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "azurerm",
		Short: "Plan Azure Resource Manager config changes",
		RunE: func(_ *cobra.Command, _ []string) error {
			PrintHeader(configDir, subscriptionId)

			if err := apply.ApplyAzureRm(configDir, subscriptionId, true); err != nil {
				return err
			}

			return nil
		},
	}

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	cmd.Flags().StringVarP(&configDir, "config-dir", "c", wd, "Config directory")
	cmd.Flags().StringVarP(&subscriptionId, "subscription-id", "s", "", "Subscription Id") // TODO: Support name

	cobra.MarkFlagRequired(cmd.Flags(), "subscription-id")

	return cmd
}

func PrintHeader(configDir string, scope string) {
	builder := &strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("~", 92)))
	builder.WriteString(fmt.Sprintf("Action           | %s\n", "Plan"))
	builder.WriteString(fmt.Sprintf("Mode             | %s\n", "Azure RM"))
	builder.WriteString(fmt.Sprintf("Config path      | %s\n", configDir))
	builder.WriteString(fmt.Sprintf("Subscription Id  | %s\n", scope))
	builder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("~", 92)))
	output.PrintlnInfo(builder.String())
}
