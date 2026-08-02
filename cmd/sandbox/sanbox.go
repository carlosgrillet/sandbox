// Package sandbox wires up and runs the sandbox CLI entrypoint.
package sandbox

import (
	"errors"
	"log/slog"
	"os"

	"github.com/carlosgrillet/sandbox/pkg/cmd"
)

// Bootstrap builds the root command from os.Args, executes it, and exits
// the process with the command's exit code (or 1 on failure).
func Bootstrap() {
	command, err := cmd.NewRootCmd(os.Stdout, os.Args[1:])
	if err != nil {
		slog.Warn("command failed", slog.Any("error", err))
		os.Exit(1)
	}

	if err := command.Execute(); err != nil {
		var commandErr cmd.CommandError
		_, errored := errors.AsType[cmd.CommandError](err)
		if errored {
			os.Exit(commandErr.ExitCode)
		}
		os.Exit(1)
	}
}
