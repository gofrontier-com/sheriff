package validate

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofrontier-com/go-utils/output"
	"github.com/gofrontier-com/sheriff/pkg/cmd/app/validate"
	"github.com/spf13/cobra"
)

var (
	configDir string
)

// NewCmdValidate creates a command to validate the Azure Rm config
func NewCmdValidateAzureRm() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "azurerm",
		Short: "Validate Azure RM config",
		RunE: func(_ *cobra.Command, _ []string) error {
			printHeader(configDir)

			if err := validate.ValidateAzureRm(configDir); err != nil {
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

	return cmd
}

func printHeader(configDir string) {
	builder := &strings.Builder{}
	builder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("~", 92)))
	builder.WriteString(fmt.Sprintf("Action       | %s\n", "Validate"))
	builder.WriteString(fmt.Sprintf("Mode         | %s\n", "Azure RM"))
	builder.WriteString(fmt.Sprintf("Config path  | %s\n", configDir))
	builder.WriteString(fmt.Sprintf("%s\n", strings.Repeat("~", 92)))
	output.PrintlnInfo(builder.String())
}
