package validate

import (
	"log"
	"os"
	"path/filepath"

	"github.com/frontierdigital/sheriff/pkg/util/validation"
	"github.com/frontierdigital/utils/output"
	"github.com/spf13/cobra"
)

// NewCmdValidate creates a command to output the current version of Sheriff
func NewCmdValidate() *cobra.Command {
	c := &cobra.Command{
		Use:   "validate groups-path",
		Short: "Validate will check input data for correctness",
		Long:  "Looks at input files and returns true or false if valid data or not",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			output.Println(args)
			entries, err := os.ReadDir(args[0])
			if err != nil {
				log.Fatal(err)
			}
			output.Println(entries)

			for _, e := range entries {
				f := filepath.Join(args[0], e.Name())
				valid, err := validation.ValidateConfiguration(f)
				if !valid {
					output.Println("Invalid %s - %v", f, err)
				} else {
					output.Println("valid %s", f)
				}
			}

			return nil
		},
	}

	return c
}
