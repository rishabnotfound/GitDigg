package version

import (
	"fmt"
	"runtime"
)

var (
	Version   = "dev"
	Commit    = "none"
	Date      = "unknown"
	GoVersion = runtime.Version()
)

func Info() string {
	return fmt.Sprintf("gitdig %s (%s, %s)", Version, Commit, Date)
}

func Short() string {
	return Version
}
