package namespace

import (
	"syscall"

	"github.com/spf13/cobra"
)

type namespace struct {
	name         string
	shorthand    string
	description  string
	defaultValue bool
	cloneFlag    uintptr
}

var namespaces []namespace

func RegisterNamespaceFlag(cmd *cobra.Command) {
	namespaces = []namespace{
		{"net", "n", "isolate network namespace", false, syscall.CLONE_NEWNET},
		{"pid", "p", "isolate PID namespace", false, syscall.CLONE_NEWPID},
		{"uts", "H", "isolate hostname namespace", false, syscall.CLONE_NEWUTS},
		{"mnt", "m", "isolate mount namespace", false, syscall.CLONE_NEWNS},
		{"ipc", "i", "isolate IPC namespace", false, syscall.CLONE_NEWIPC},
		{"user", "u", "isolate user namespace", false, syscall.CLONE_NEWUSER},
		{"time", "t", "isolate time namespace", false, syscall.CLONE_NEWTIME},
		{"cgroup", "c", "isolate cgroup namespace", false, syscall.CLONE_NEWCGROUP},
	}

	for _, ns := range namespaces {
		cmd.Flags().BoolP(ns.name, ns.shorthand, ns.defaultValue, ns.description)
	}
}

func GetCloneFlags(cmd *cobra.Command) uintptr {
	var flagsSum uintptr
	for _, ns := range namespaces {
		// TODO: catch the error here
		flag, _ := cmd.Flags().GetBool(ns.name)
		if flag {
			flagsSum = flagsSum | ns.cloneFlag
		}
	}

	return flagsSum
}
