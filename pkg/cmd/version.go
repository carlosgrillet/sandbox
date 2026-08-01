package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/carlosgrillet/sandbox/internal/version"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "print the sb version information",
		RunE:  runVersion,
	}
}

func runVersion(_ *cobra.Command, _ []string) error {
	v := version.Get()
	out := v.Version
	if len(v.GitCommit) >= 7 {
		out += "+" + v.GitCommit[:7]
	}
	fmt.Println(out)
	return nil
}
