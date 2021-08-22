package common

import (
	"fmt"
	"time"
)

var (
	// Version should be a git tag
	Version = ""
	// BuildTime in example: 2019-06-29T23:23:09+0800
	BuildTime = ""
	UpTime    time.Time
	Branch    = ""
	Commit    = ""
)

func PrintVersion() string {
	return fmt.Sprintf("%s (git: %s %s)\n", Version, Branch, Commit)
}
