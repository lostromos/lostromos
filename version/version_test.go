package version

func ExamplePrint() {
	Version = "1"
	GitHash = "abc123"
	BuildTime = "Some point in time"

	Print()
	// Output:
	// Version: 1
	// Git Commit Hash: abc123
	// Build Time: Some point in time
}
