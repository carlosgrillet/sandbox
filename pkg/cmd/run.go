package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/carlosgrillet/sandbox/internal/namespace"
	"github.com/carlosgrillet/sandbox/internal/rootfs"
)

// newRunCommand builds the "run" sub-command, which executes a process in
// the namespaces selected by --net/--pid/--uts/... (or --all).
func newRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [flags] EXECUTABLE",
		Short: "run a new process with isolation on the selected namespaces",
		Args:  cobra.MinimumNArgs(1),
		RunE:  run,
	}

	namespace.RegisterNamespaceFlag(cmd)

	cmd.Flags().BoolP("all", "A", false, "isolate all namespaces")
	cmd.Flags().SetInterspersed(false)

	return cmd
}

// run is the RunE for "run". If the mnt namespace is requested it
// materializes a rootfs and re-execs "/proc/self/exe __init ..." into it
// (a new mount namespace only takes effect for a new process, not the
// caller); otherwise it execs args[0] directly with the requested clone
// flags.
func run(cmd *cobra.Command, args []string) (runErr error) {
	mountNamespace, err := namespaceEnabled(cmd, "mnt")
	if err != nil {
		return err
	}

	sysProcAttr, err := getSysProcAttr(cmd)
	if err != nil {
		return err
	}

	if mountNamespace {
		var filesystem *rootfs.RootFS
		filesystem, err = rootfs.Materialize()
		if err != nil {
			return err
		}
		defer func() {
			runErr = errors.Join(runErr, filesystem.Close())
		}()

		var mountProc bool
		mountProc, err = namespaceEnabled(cmd, "pid")
		if err != nil {
			return err
		}
		initArgs := []string{
			"__init",
			filesystem.Path(),
			strconv.FormatBool(mountProc),
		}
		initArgs = append(initArgs, args...)

		command := exec.CommandContext(cmd.Context(), "/proc/self/exe", initArgs...)
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		command.SysProcAttr = sysProcAttr
		return command.Run()
	}

	commandPath, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("find executable %q: %w", args[0], err)
	}

	command := &exec.Cmd{
		Path:        commandPath,
		Args:        args,
		Stdin:       os.Stdin,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		SysProcAttr: sysProcAttr,
	}

	return command.Run()
}

// getSysProcAttr builds the SysProcAttr for the child process: clone flags
// for the requested namespaces, plus a uid/gid mapping (root inside, current
// user outside) when the user namespace is requested via --user or --all.
func getSysProcAttr(cmd *cobra.Command) (*syscall.SysProcAttr, error) {
	cloneFlags, err := namespace.GetCloneFlags(cmd)
	if err != nil {
		return nil, err
	}

	procAttr := &syscall.SysProcAttr{
		Cloneflags: cloneFlags,
	}

	user, err := cmd.Flags().GetBool("user")
	if err != nil {
		return nil, fmt.Errorf("get user flag: %w", err)
	}
	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return nil, fmt.Errorf("get all flag: %w", err)
	}

	if user || all {
		procAttr.UidMappings = []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		}
		procAttr.GidMappings = []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		}
		procAttr.GidMappingsEnableSetgroups = false
	}
	return procAttr, nil
}

// namespaceEnabled reports whether the named namespace flag is set, treating
// --all as enabling every namespace regardless of its individual flag.
func namespaceEnabled(cmd *cobra.Command, name string) (bool, error) {
	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return false, fmt.Errorf("get all flag: %w", err)
	}
	if all {
		return true, nil
	}

	enabled, err := cmd.Flags().GetBool(name)
	if err != nil {
		return false, fmt.Errorf("get %s flag: %w", name, err)
	}
	return enabled, nil
}
