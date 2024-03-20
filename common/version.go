package common

import (
	"runtime/debug"
)

func GetVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		version := "azukiiro/"
		if info.Main.Version == "(devel)" {
			version += "devel"
		} else {
			version += info.Main.Version
		}
		version += " ("
		version += info.Main.Path
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				version += "; " + setting.Value[:7]
			case "vcs.modified":
				if setting.Value == "true" {
					version += "; modified"
				}
			}
		}
		version += ")"
		return version
	}
	return "azukiiro/unknown"
}
