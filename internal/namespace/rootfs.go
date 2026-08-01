package namespace

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// oldRootName is the directory used as pivot_root's put_old target. It is
// created under the new root and removed once the old root is detached.
const oldRootName = ".oldroot"

// EnterRootFS replaces the process root with root inside the current mount
// namespace. The caller must already be in a private mount namespace
// (e.g. after CLONE_NEWNS / unshare(CLONE_NEWNS)).
func EnterRootFS(root string, mountProc bool) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve rootfs: %w", err)
	}

	if err := validateDir(root); err != nil {
		return err
	}
	if err := privatizeMountPropagation(); err != nil {
		return err
	}
	if err := bindMountSelf(root); err != nil {
		return err
	}
	if err := pivotRoot(root); err != nil {
		return err
	}
	// The kernel refuses to mount a fresh procfs once the old root has been
	// detached, so mount it first and detach the old root afterwards.
	if mountProc {
		if err := mountProcfs(); err != nil {
			return err
		}
	}
	return detachOldRoot()
}

// validateDir checks that root exists and is a directory.
func validateDir(root string) error {
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("inspect rootfs: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("rootfs %q is not a directory", root)
	}
	return nil
}

// privatizeMountPropagation ensures mount/unmount events in this namespace
// never propagate to the host, and vice versa. Required before pivot_root.
func privatizeMountPropagation() error {
	if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("make mount propagation private: %w", err)
	}
	return nil
}

// bindMountSelf bind-mounts root onto itself so it satisfies pivot_root's
// requirement that new_root be a mount point.
func bindMountSelf(root string) error {
	if err := unix.Mount(root, root, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
		return fmt.Errorf("bind mount rootfs: %w", err)
	}
	return nil
}

// pivotRoot swaps the process root to root. The old root is left mounted at
// oldRootName under the new root; call detachOldRoot to remove it once any
// namespace-scoped filesystems (e.g. procfs) have been mounted.
func pivotRoot(root string) error {
	oldRoot := filepath.Join(root, oldRootName)
	if err := os.Mkdir(oldRoot, 0o700); err != nil {
		return fmt.Errorf("create old root directory: %w", err)
	}
	if err := unix.PivotRoot(root, oldRoot); err != nil {
		return fmt.Errorf("pivot root: %w", err)
	}
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("change to new root: %w", err)
	}
	return nil
}

// detachOldRoot unmounts and removes the old root left behind by pivotRoot.
func detachOldRoot() error {
	oldRootAbs := "/" + oldRootName
	if err := unix.Unmount(oldRootAbs, unix.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount old root: %w", err)
	}
	if err := os.Remove(oldRootAbs); err != nil {
		return fmt.Errorf("remove old root directory: %w", err)
	}
	return nil
}

// mountProcfs mounts a fresh procfs at /proc, scoped to the new namespace.
func mountProcfs() error {
	if err := os.MkdirAll("/proc", 0o555); err != nil {
		return fmt.Errorf("create /proc: %w", err)
	}
	if err := unix.Mount(
		"proc",
		"/proc",
		"proc",
		unix.MS_NOSUID|unix.MS_NOEXEC|unix.MS_NODEV,
		"",
	); err != nil {
		return fmt.Errorf("mount /proc: %w", err)
	}
	return nil
}
