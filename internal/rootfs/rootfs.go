package rootfs

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

//go:embed assets/alpine-3.23.5.tar.gz
var alpineArchive []byte

// RootFS is a materialized copy of the embedded root filesystem.
// Call Close after the process using it has exited.
type RootFS struct {
	path string
}

// Materialize extracts the embedded root filesystem into a new temporary
// directory. Each call returns an independent, writable filesystem.
func Materialize() (*RootFS, error) {
	directory, err := os.MkdirTemp("", "sandbox-rootfs-")
	if err != nil {
		return nil, fmt.Errorf("create rootfs directory: %w", err)
	}

	root := &RootFS{path: directory}
	if err := extractGzipTar(directory, alpineArchive); err != nil {
		cleanupErr := root.Close()
		return nil, errors.Join(fmt.Errorf("extract rootfs: %w", err), cleanupErr)
	}

	return root, nil
}

// Path returns the directory containing the materialized root filesystem.
func (r *RootFS) Path() string {
	return r.path
}

// Close removes the materialized root filesystem.
func (r *RootFS) Close() error {
	if r == nil || r.path == "" {
		return nil
	}

	directory := r.path
	r.path = ""
	if err := os.RemoveAll(directory); err != nil {
		return fmt.Errorf("remove rootfs %q: %w", directory, err)
	}
	return nil
}

func extractGzipTar(root string, archive []byte) error {
	gzipReader, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return fmt.Errorf("open gzip archive: %w", err)
	}
	defer gzipReader.Close()

	return extractTar(root, tar.NewReader(gzipReader))
}

type directoryMode struct {
	path string
	mode os.FileMode
}

type hardlink struct {
	path   string
	target string
}

func extractTar(root string, archive *tar.Reader) error {
	var directories []directoryMode
	var hardlinks []hardlink

	for {
		header, err := archive.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar header: %w", err)
		}

		target, err := archivePath(root, header.Name)
		if err != nil {
			return err
		}

		if err := makeParents(root, filepath.Dir(target)); err != nil {
			return fmt.Errorf("prepare parent for %q: %w", header.Name, err)
		}

		mode := header.FileInfo().Mode()

		switch header.Typeflag {
		case tar.TypeDir:
			if err := makeDirectory(target); err != nil {
				return fmt.Errorf("create directory %q: %w", header.Name, err)
			}
			// Apply restrictive directory modes after all children are extracted.
			directories = append(directories, directoryMode{target, mode})

		case tar.TypeReg:
			if err := writeFile(target, mode, archive); err != nil {
				return fmt.Errorf("extract file %q: %w", header.Name, err)
			}

		case tar.TypeSymlink:
			if err := os.Symlink(header.Linkname, target); err != nil {
				return fmt.Errorf("create symlink %q: %w", header.Name, err)
			}

		case tar.TypeLink:
			linkTarget, err := archivePath(root, header.Linkname)
			if err != nil {
				return fmt.Errorf("invalid hardlink target for %q: %w", header.Name, err)
			}
			hardlinks = append(hardlinks, hardlink{target, linkTarget})

		case tar.TypeXGlobalHeader:
			// archive/tar handles the global PAX metadata for subsequent entries.

		default:
			return fmt.Errorf("unsupported tar entry %q with type %d", header.Name, header.Typeflag)
		}
	}

	for _, link := range hardlinks {
		if err := makeParents(root, filepath.Dir(link.path)); err != nil {
			return fmt.Errorf("prepare hardlink parent: %w", err)
		}
		info, err := os.Lstat(link.target)
		if err != nil {
			return fmt.Errorf("inspect hardlink target %q: %w", link.target, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("hardlink target %q is not a regular file", link.target)
		}
		if err := os.Link(link.target, link.path); err != nil {
			return fmt.Errorf("create hardlink %q: %w", link.path, err)
		}
	}

	// Children are applied before parents so a restrictive parent does not
	// prevent chmod calls below it.
	for _, directory := range slices.Backward(directories) {
		if err := os.Chmod(directory.path, directory.mode); err != nil {
			return fmt.Errorf("set directory mode %q: %w", directory.path, err)
		}
	}

	return nil
}

func archivePath(root, name string) (string, error) {
	if path.IsAbs(name) {
		return "", fmt.Errorf("archive path %q is absolute", name)
	}

	cleanName := path.Clean(name)
	if cleanName == ".." || strings.HasPrefix(cleanName, "../") {
		return "", fmt.Errorf("archive path %q escapes the rootfs", name)
	}

	target := filepath.Join(root, filepath.FromSlash(cleanName))
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return "", fmt.Errorf("resolve archive path %q: %w", name, err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("archive path %q escapes the rootfs", name)
	}
	return target, nil
}

func makeParents(root, parent string) error {
	relative, err := filepath.Rel(root, parent)
	if err != nil {
		return err
	}
	if relative == "." {
		return nil
	}

	current := root
	for component := range strings.SplitSeq(relative, string(filepath.Separator)) {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		switch {
		case errors.Is(err, os.ErrNotExist):
			if err := os.Mkdir(current, 0o755); err != nil {
				return err
			}
		case err != nil:
			return err
		case !info.IsDir():
			return fmt.Errorf("path component %q is not a directory", current)
		}
	}
	return nil
}

func makeDirectory(target string) error {
	info, err := os.Lstat(target)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return os.Mkdir(target, 0o755)
	case err != nil:
		return err
	case !info.IsDir():
		return errors.New("path already exists and is not a directory")
	default:
		return nil
	}
}

func writeFile(target string, mode os.FileMode, contents io.Reader) error {
	file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(file, contents)
	closeErr := file.Close()
	if copyErr != nil || closeErr != nil {
		return errors.Join(copyErr, closeErr)
	}
	return os.Chmod(target, mode)
}
