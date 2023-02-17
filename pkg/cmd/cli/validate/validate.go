package validate

import (
	"github.com/frontierdigital/sheriff/pkg/util/validation"
	"github.com/spf13/cobra"
	// 	hashiVers "github.com/hashicorp/go-version"
	// 	"github.com/hashicorp/hc-install/product"
	// 	"github.com/hashicorp/hc-install/releases"
	// 	"github.com/hashicorp/terraform-exec/tfexec"
)

var (
	inputDir = ""
)

// NewCmdValidate creates a command to output the current version of Sheriff
func NewCmdValidate() *cobra.Command {
	c := &cobra.Command{
		Use:   "validate",
		Short: "Validate will check input data for correctness",
		Long:  "Looks at input files and returns true or false if valid data or not",
		RunE: func(_ *cobra.Command, _ []string) error {

			validation.ValidateConfiguration("./sre.yml")

			return nil
		},
	}

	c.Flags().StringVarP(&inputDir, "input-dir", "i", "", "Directory holding the input.")

	return c
}
