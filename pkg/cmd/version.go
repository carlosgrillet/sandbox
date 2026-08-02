package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/carlosgrillet/sandbox/internal/version"
)

// newVersionCommand builds the "version" sub-command.
func newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "print the sb version information",
		RunE:  runVersion,
	}

	cmd.Flags().Bool("short", false, "print short version")

	return cmd
}

// runVersion is the RunE for "version": prints the version, suffixed with
// the short (7-char) git commit hash when available.
func runVersion(cmd *cobra.Command, _ []string) error {
	versionInfo := version.Get()

	short, err := cmd.Flags().GetBool("short")
	if err != nil {
		return err
	}

	if short {
		fmt.Println(versionInfo.Version + "+" + versionInfo.GitCommit[:7])
		return nil
	}

	fmt.Printf("%#v\n", versionInfo)
	return nil
}
