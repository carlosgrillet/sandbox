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
		Short: "run a new process",
		RunE:  run,
	}

	namespace.RegisterNamespaceFlag(cmd)

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


	command := &exec.Cmd{
		Path:        commandPath,
		Args:        args,
		Stdin:       os.Stdin,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		SysProcAttr: getSysProcAttr(cmd),
	}

	return command.Run()
}

func getSysProcAttr(cmd *cobra.Command) *syscall.SysProcAttr {
	procAttr := &syscall.SysProcAttr{
		Cloneflags: namespace.GetCloneFlags(cmd),
	}

	// TODO: catch the error here
	flag, _ := cmd.Flags().GetBool("user")
	if flag {
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
	return procAttr
}
