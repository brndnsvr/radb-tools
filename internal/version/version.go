// Package version provides centralized version information for the application.
package version

import (
	"fmt"
	"runtime"
)

// These variables are set at build time via -ldflags
var (
	// Version is the semantic version of the application
	Version = "0.0.43"

	// GitCommit is the git commit hash (short)
	GitCommit = "dev"

	// GitBranch is the git branch
	GitBranch = "unknown"

	// BuildDate is the date the binary was built
	BuildDate = "unknown"

	// GoVersion is the version of Go used to build
	GoVersion = runtime.Version()

	// Platform is the OS/Arch combination
	Platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)

// Info contains all version information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	GitBranch string `json:"git_branch"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// Get returns the complete version information
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		GitBranch: GitBranch,
		BuildDate: BuildDate,
		GoVersion: GoVersion,
		Platform:  Platform,
	}
}

// String returns a formatted version string
func String() string {
	return fmt.Sprintf("radb-client version %s (commit: %s, built: %s)",
		Version, GitCommit, BuildDate)
}

// Short returns just the version number
func Short() string {
	return Version
}

// Full returns a detailed multi-line version string
func Full() string {
	return fmt.Sprintf(`radb-client version %s

Build Information:
  Git Commit:   %s
  Git Branch:   %s
  Build Date:   %s
  Go Version:   %s
  Platform:     %s`,
		Version,
		GitCommit,
		GitBranch,
		BuildDate,
		GoVersion,
		Platform,
	)
}

// IsPreRelease returns true if this is a pre-release version
func IsPreRelease() bool {
	return GitCommit == "dev" ||
		   contains(Version, "-pre") ||
		   contains(Version, "-alpha") ||
		   contains(Version, "-beta") ||
		   contains(Version, "-rc")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		   (s == substr ||
		    (len(s) > len(substr) &&
		     s[len(s)-len(substr):] == substr))
}
