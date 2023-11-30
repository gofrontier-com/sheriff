package version

import (
	"github.com/frontierdigital/utils/output"
	"github.com/spf13/cobra"
	goVersion "go.hein.dev/go-version"
)

var (
	outputFmt string
	shortened bool
)

// NewCmdVersion creates a command to output the current version
func NewCmdVersion(version string, commit string, date string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Output version information",
		RunE: func(_ *cobra.Command, _ []string) error {
			resp := goVersion.FuncWithOutput(shortened, version, commit, date, outputFmt)
			output.PrintfInfo(resp)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&shortened, "short", "s", false, "Print just the version number")
	cmd.Flags().StringVarP(&outputFmt, "output", "o", "json", "Output format")

	return cmd
}
