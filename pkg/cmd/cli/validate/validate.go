package validate

import (
	"github.com/frontierdigital/sheriff/pkg/cmd/app/validate"
	"github.com/spf13/cobra"
)

// NewCmdValidate creates a command to output the current version of Sheriff
func NewCmdValidate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate configuration files",
		Args:  cobra.ExactArgs(0),
		RunE: func(_ *cobra.Command, _ []string) error {
			// return validate.Validate()
			// output.Println(args)
			// entries, err := os.ReadDir(args[0])
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// output.Println(entries)

			// for _, e := range entries {
			// 	f := filepath.Join(args[0], e.Name())
			// 	valid, err := validation.ValidateConfiguration(f)
			// 	if !valid {
			// 		output.Println("Invalid %s - %v", f, err)
			// 	} else {
			// 		output.Println("valid %s", f)
			// 	}
			// }

			// return nil
			if err := validate.Validate(); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
