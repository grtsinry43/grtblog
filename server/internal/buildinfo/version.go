package buildinfo

import (
	"runtime/debug"
	"strings"
)

// BuildVersion can be injected at build time via:
// -ldflags "-X github.com/grtsinry43/grtblog-v2/server/internal/buildinfo.BuildVersion=<value>"
var BuildVersion string

// Version resolves the build/version string embedded in the binary.
func Version() string {
	if v := strings.TrimSpace(BuildVersion); v != "" {
		return v
	}

	info, ok := debug.ReadBuildInfo()
	if !ok || info == nil {
		return "dev"
	}
	if info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}

	var revision string
	var modified string
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			modified = setting.Value
		}
	}
	if revision != "" {
		short := shortCommit(revision)
		if modified == "true" {
			return short + "-dirty"
		}
		return short
	}
	if info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}

func shortCommit(revision string) string {
	rev := strings.TrimSpace(revision)
	if len(rev) > 12 {
		return rev[:12]
	}
	if rev == "" {
		return "dev"
	}
	return rev
}
