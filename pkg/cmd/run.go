package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/carlosgrillet/sandbox/internal/namespace"
	"github.com/spf13/cobra"
)

func newRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [flags] EXECUTABLE",
		Short: "run a new process with isolation on the selected namespaces",
		RunE:  run,
	}

	namespace.RegisterNamespaceFlag(cmd)

	cmd.Flags().BoolP("all", "A", false, "isolate all namespaces")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	commandPath, err := exec.LookPath(args[0])
	if err != nil {
		return fmt.Errorf("can't find executable %q. %v", commandPath, err)
	}

	// TODO: when passing mnt
	//   - obtain the rootfs (static alpine image? OCI images integration?)
	//   - spawn child in new namespaces
	//   Child:
	//   - make mount propagation private
	//   - mount and bind the rootfs
	//   - pivot_root
	//   - chdir to /
	//   - unmount the old host root
	//   - mount /proc for the new PID namespace
	//   - execute the requested command

	sysProcAttr, err := getSysProcAttr(cmd)
	if err != nil {
		return err
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

func getSysProcAttr(cmd *cobra.Command) (*syscall.SysProcAttr, error) {
	cloneFlags, err := namespace.GetCloneFlags(cmd)
	if err != nil {
		return nil, err
	}

	procAttr := &syscall.SysProcAttr{
		Cloneflags: cloneFlags,
	}

	// TODO: catch the error here
	user, _ := cmd.Flags().GetBool("user")
	all, _ := cmd.Flags().GetBool("all")

	if user || all {
		procAttr.UidMappings = []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      1000,
				Size:        1,
			},
		}
		procAttr.GidMappings = []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      1000,
				Size:        1,
			},
		}
	}
	return procAttr, nil
}
