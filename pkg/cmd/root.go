// Package cmd builds the sandbox cobra command tree.
package cmd

import (
	"io"

	"github.com/spf13/cobra"
)

// CommandError wraps an error with the process exit code it should produce.
type CommandError struct {
	error
	ExitCode int
}

// NewRootCmd builds the "sandbox" root command and registers its sub-commands.
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
		newVersionCommand(),
	)

	return cmd, nil
}
