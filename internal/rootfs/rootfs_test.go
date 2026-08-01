package rootfs

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaterialize(t *testing.T) {
	filesystem, err := Materialize()
	if err != nil {
		t.Fatal(err)
	}
	root := filesystem.Path()

	busybox, err := os.Stat(filepath.Join(root, "bin", "busybox"))
	if err != nil {
		t.Fatal(err)
	}
	if !busybox.Mode().IsRegular() || busybox.Mode().Perm()&0o111 == 0 {
		t.Fatalf("busybox has unexpected mode %v", busybox.Mode())
	}

	link, err := os.Readlink(filepath.Join(root, "bin", "sh"))
	if err != nil {
		t.Fatal(err)
	}
	if link != "/bin/busybox" {
		t.Fatalf("/bin/sh points to %q, want /bin/busybox", link)
	}

	if err := filesystem.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("rootfs still exists after Close: %v", err)
	}
	if err := filesystem.Close(); err != nil {
		t.Fatalf("second Close returned an error: %v", err)
	}
}

func TestExtractTarPreservesModesAndHardlinks(t *testing.T) {
	archive := makeTar(t,
		tarEntry{name: "bin", kind: tar.TypeDir, mode: 0o750},
		tarEntry{name: "bin/tool", kind: tar.TypeReg, mode: 0o751, contents: "tool"},
		tarEntry{name: "bin/tool-link", kind: tar.TypeLink, link: "bin/tool"},
	)
	root := t.TempDir()

	if err := extractTar(root, tar.NewReader(bytes.NewReader(archive))); err != nil {
		t.Fatal(err)
	}

	directory, err := os.Stat(filepath.Join(root, "bin"))
	if err != nil {
		t.Fatal(err)
	}
	if got := directory.Mode().Perm(); got != 0o750 {
		t.Fatalf("directory mode = %o, want 750", got)
	}

	target, err := os.Stat(filepath.Join(root, "bin", "tool"))
	if err != nil {
		t.Fatal(err)
	}
	link, err := os.Stat(filepath.Join(root, "bin", "tool-link"))
	if err != nil {
		t.Fatal(err)
	}
	if got := target.Mode().Perm(); got != 0o751 {
		t.Fatalf("file mode = %o, want 751", got)
	}
	if !os.SameFile(target, link) {
		t.Fatal("hardlink does not reference the same file")
	}
}

func TestExtractTarRejectsEscapingPaths(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "absolute", path: "/outside"},
		{name: "parent traversal", path: "../../outside"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			archive := makeTar(t, tarEntry{
				name:     test.path,
				kind:     tar.TypeReg,
				mode:     0o644,
				contents: "bad",
			})

			err := extractTar(t.TempDir(), tar.NewReader(bytes.NewReader(archive)))
			if err == nil || !strings.Contains(err.Error(), "rootfs") && !strings.Contains(err.Error(), "absolute") {
				t.Fatalf("expected unsafe path error, got %v", err)
			}
		})
	}
}

func TestExtractTarRejectsSymlinkParent(t *testing.T) {
	outside := t.TempDir()
	archive := makeTar(t,
		tarEntry{name: "escape", kind: tar.TypeSymlink, link: outside},
		tarEntry{name: "escape/file", kind: tar.TypeReg, mode: 0o644, contents: "bad"},
	)

	err := extractTar(t.TempDir(), tar.NewReader(bytes.NewReader(archive)))
	if err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("expected symlink parent error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(outside, "file")); !os.IsNotExist(err) {
		t.Fatalf("archive wrote through symlink outside rootfs: %v", err)
	}
}

type tarEntry struct {
	name     string
	kind     byte
	mode     int64
	link     string
	contents string
}

func makeTar(t *testing.T, entries ...tarEntry) []byte {
	t.Helper()

	var buffer bytes.Buffer
	writer := tar.NewWriter(&buffer)
	for _, entry := range entries {
		header := &tar.Header{
			Name:     entry.name,
			Typeflag: entry.kind,
			Mode:     entry.mode,
			Linkname: entry.link,
			Size:     int64(len(entry.contents)),
		}
		if err := writer.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if _, err := writer.Write([]byte(entry.contents)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}
