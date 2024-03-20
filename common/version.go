package common

import (
	"runtime/debug"
)

var (
	Version string
)

func init() {
	Version = "azukiiro/"
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version == "(devel)" {
			Version += "devel"
		} else {
			Version += info.Main.Version
		}
		Version += " ("
		Version += info.Main.Path
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				Version += "; " + setting.Value[:7]
			case "vcs.modified":
				if setting.Value == "true" {
					Version += "; modified"
				}
			}
		}
		Version += ")"
	} else {
		Version += "unknown"
	}
}

func GetVersion() string {
	return Version
}
