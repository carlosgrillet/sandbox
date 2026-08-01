package cmd

import (
	"io"

	"github.com/spf13/cobra"
)

type CommandError struct {
	error
	ExitCode int
}

func NewRootCmd(_ io.Writer, args []string) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:          "sb",
		Short:        "A lightweight tool for running commands in isolated Linux namespaces",
		SilenceUsage: true,
	}

	flags := cmd.PersistentFlags()
	flags.Parse(args)

	// Sub-commands registration
	cmd.AddCommand(
		newInitCommand(),
		newRunCommand(),
	)

	return cmd, nil
}
