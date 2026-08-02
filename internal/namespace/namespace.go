// Package namespace registers per-namespace cobra flags (--net, --pid, ...)
// and translates the cobra flags into Linux clone(2) namespace flags.
package namespace

import (
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
)

// namespace describes one Linux namespace exposed as a CLI flag.
type namespace struct {
	name         string
	shorthand    string
	description  string
	defaultValue bool
	cloneFlag    uintptr
}

// namespaces holds the set registered by RegisterNamespaceFlag; GetCloneFlags
// reads it back, so RegisterNamespaceFlag must run first.
var namespaces []namespace

// RegisterNamespaceFlag adds one bool flag per supported namespace
// (net, pid, uts, mnt, ipc, user, time, cgroup) to cmd, and populates the
// package-level namespace table used by GetCloneFlags.
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

// GetCloneFlags reads the flags registered by RegisterNamespaceFlag off cmd
// and ORs together the clone(2) flags for every enabled namespace. If the
// "all" flag is set, every namespace is enabled regardless of its own flag.
func GetCloneFlags(cmd *cobra.Command) (uintptr, error) {
	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return 0, fmt.Errorf("get all flag: %w", err)
	}

	var cloneFlags uintptr

	for _, ns := range namespaces {
		enabled := all

		if !all {
			enabled, err = cmd.Flags().GetBool(ns.name)
			if err != nil {
				return 0, fmt.Errorf("get %s flag: %w", ns.name, err)
			}
		}

		if enabled {
			cloneFlags |= ns.cloneFlag
		}
	}

	return cloneFlags, nil
}
