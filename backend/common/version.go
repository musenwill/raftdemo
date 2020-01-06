package common

var (
	// AppName is executable binary file name
	AppName = ""
	// Version should be a git tag
	Version = ""
	// BuildTime in example: 2019-06-29T23:23:09+0800
	BuildTime = ""
)

// check Version and BuildTime
func init() {
	if len(AppName) <= 0 {
		panic("AppName unset, expect to be executable binary file name")
	}
	if len(Version) <= 0 {
		panic("Version unset, expect be set with a git tag")
	}
	if len(BuildTime) <= 0 {
		panic("BuildTime unset, expect to be like '2019-06-29T23:23:09+0800'")
	}
}
