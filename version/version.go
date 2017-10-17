package version

import (
	"fmt"
)

// These variables should be substituted with real values during build
var (
	Version   = ""
	GitHash   = ""
	BuildTime = ""
)

// Print prints version information to stdout
func Print() {
	fmt.Println("Version:", Version)
	fmt.Println("Git Commit Hash:", GitHash)
	fmt.Println("Build Time:", BuildTime)
}
