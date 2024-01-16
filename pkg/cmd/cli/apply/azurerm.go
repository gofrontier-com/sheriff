package apply

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
	planOnly       bool
	subscriptionId string
)

// NewCmdApplyAzureRm creates a command to apply the Azure RM config
func NewCmdApplyAzureRm() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "azurerm",
		Short: "Apply Azure Resource Manager config",
		RunE: func(_ *cobra.Command, _ []string) error {
			printHeader(configDir, subscriptionId)

			if err := apply.ApplyAzureRm(configDir, subscriptionId, planOnly); err != nil {
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
	cmd.Flags().BoolVarP(&planOnly, "plan-only", "p", false, "Plan-only")
	cmd.Flags().StringVarP(&subscriptionId, "subscription-id", "s", "", "Subscription Id") // TODO: Support name

	cobra.MarkFlagRequired(cmd.Flags(), "subscription-id")

	return cmd
}

func printHeader(configDir string, scope string) {
	var action string
	if planOnly {
		action = "Apply (plan-only)"
	} else {
		action = "Apply"
	}

	builder := &strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("~", 92)))
	builder.WriteString(fmt.Sprintf("Action           | %s\n", action))
	builder.WriteString(fmt.Sprintf("Mode             | %s\n", "Azure RM"))
	builder.WriteString(fmt.Sprintf("Config path      | %s\n", configDir))
	builder.WriteString(fmt.Sprintf("Subscription Id  | %s\n", scope))
	builder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("~", 92)))
	output.PrintlnInfo(builder.String())
}
