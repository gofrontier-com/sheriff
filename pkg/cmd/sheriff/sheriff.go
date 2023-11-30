package sheriff

import (
	"os"

	"github.com/frontierdigital/sheriff/pkg/cmd/cli/apply"
	"github.com/frontierdigital/sheriff/pkg/cmd/cli/validate"
	vers "github.com/frontierdigital/sheriff/pkg/cmd/cli/version"
	"github.com/frontierdigital/sheriff/pkg/util/app_config"
	"github.com/frontierdigital/utils/output"
	"github.com/spf13/cobra"
)

func NewRootCmd(version string, commit string, date string) *cobra.Command {
	_, err := app_config.LoadAppConfig()
	if err != nil {
		output.PrintlnError(err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:                   "sheriff",
		DisableFlagsInUseLine: true,
		Short:                 "sheriff is the command line tool for Sheriff",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Help(); err != nil {
				return err
			}

			return nil
		},
	}

	rootCmd.AddCommand(apply.NewCmdApply())
	rootCmd.AddCommand(validate.NewCmdValidate())
	rootCmd.AddCommand(vers.NewCmdVersion(version, commit, date))

	return rootCmd
}
