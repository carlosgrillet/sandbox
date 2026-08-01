package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/carlosgrillet/sandbox/internal/namespace"
)

const rootfsPath = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:                "__init ROOTFS MOUNT_PROC EXECUTABLE [ARG...]",
		Hidden:             true,
		DisableFlagParsing: true,
		Args:               cobra.MinimumNArgs(3),
		RunE:               runInit,
	}
}

func runInit(_ *cobra.Command, args []string) error {
	mountProc, err := strconv.ParseBool(args[1])
	if err != nil {
		return fmt.Errorf("parse mount-proc value: %w", err)
	}

	if err := namespace.EnterRootFS(args[0], mountProc); err != nil {
		return err
	}
	if err = os.Setenv("PATH", rootfsPath); err != nil {
		return fmt.Errorf("set rootfs PATH: %w", err)
	}

	commandArgs := args[2:]
	commandPath, err := exec.LookPath(commandArgs[0])
	if err != nil {
		return fmt.Errorf("find executable %q in rootfs: %w", commandArgs[0], err)
	}

	if err := syscall.Exec(commandPath, commandArgs, os.Environ()); err != nil {
		return fmt.Errorf("execute %q: %w", commandPath, err)
	}
	return nil
}
