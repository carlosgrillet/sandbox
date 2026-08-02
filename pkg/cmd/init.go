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

// rootfsPath is the PATH set inside the materialized rootfs, since the
// host's PATH doesn't apply once pivot_root has switched filesystems.
const rootfsPath = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

// newInitCommand builds the hidden "__init" sub-command. It's not meant to
// be invoked by users directly; "run" re-execs the binary with this
// sub-command (via /proc/self/exe) to complete setup inside new namespaces.
// See run.go's run function for why the re-exec is needed.
func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:                "__init ROOTFS MOUNT_PROC EXECUTABLE [ARG...]",
		Hidden:             true,
		DisableFlagParsing: true,
		Args:               cobra.MinimumNArgs(3),
		RunE:               runInit,
	}
}

// runInit is the RunE for "__init". args is [rootfs, mountProc, executable,
// arg...]: it enters the given rootfs via namespace.EnterRootFS, resets
// PATH for the new filesystem, then replaces this process (syscall.Exec)
// with the requested executable.
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
