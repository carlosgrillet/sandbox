package cmd

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func newRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run EXECUTABLE",
		Short: "run a new process",
		RunE:  run,
	}

	return cmd
}

func run(_ *cobra.Command, args []string) error {
	command := &exec.Cmd{
		Path:   args[0],
		Args:   args,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	return command.Run()
}
