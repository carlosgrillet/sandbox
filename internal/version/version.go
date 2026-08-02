// Package version exposes build-time version info. The var values below are
// overwritten via -ldflags -X at build time (see .goreleaser.yaml); they
// default to "dev"/"" for local, non-release builds.
package version

import (
	"runtime"
)

var (
	version      = "dev"
	metadata     = ""
	gitCommit    = ""
	gitTreeState = ""
)

// BuildInfo is the JSON-serializable snapshot returned by Get.
type BuildInfo struct {
	Version      string `json:"version,omitempty"`
	GitCommit    string `json:"git_commit,omitempty"`
	GitTreeState string `json:"git_tree_state,omitempty"`
	GoVersion    string `json:"go_version,omitempty"`
}

// GetVersion returns the semantic version, suffixed with "+metadata" when
// build metadata was injected at build time.
func GetVersion() string {
	if metadata == "" {
		return version
	}
	return version + "+" + metadata
}

// Get returns the full build info: version, git commit, git tree state
// (clean/dirty), and the Go runtime version used to build the binary.
func Get() BuildInfo {
	return BuildInfo{
		Version:      GetVersion(),
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		GoVersion:    runtime.Version(),
	}
}
