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

type BuildInfo struct {
	Version      string `json:"version,omitempty"`
	GitCommit    string `json:"git_commit,omitempty"`
	GitTreeState string `json:"git_tree_state,omitempty"`
	GoVersion    string `json:"go_version,omitempty"`
}

func GetVersion() string {
	if metadata == "" {
		return version
	}
	return version + "+" + metadata
}

func Get() BuildInfo {
	return BuildInfo{
		Version:      GetVersion(),
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		GoVersion:    runtime.Version(),
	}
}
