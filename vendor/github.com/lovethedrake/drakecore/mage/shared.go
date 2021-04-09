package mage

import (
	"fmt"
	"os"

	"github.com/carolynvs/magex/shx"
)

var outDir = ""

// Any commands executed by "must" (as opposed to shx.RunV for example), will stop
// the build immediately when the command fails.
var must = shx.CommandBuilder{StopOnError: true}

// Figure out where output files from the build should be placed
func GetOutputDir() string {
	if outDir != "" {
		return outDir
	}
	const sharedVolume = "/shared"
	if _, err := os.Stat(sharedVolume); err == nil {
		return sharedVolume
	}

	return "."
}

// Sync dependencies defined in go.mod with vendor/
func Vendor() {
	must.RunV("go", "mod", "tidy")
	must.RunV("go", "mod", "vendor")
}

// Check if go.mod matches the contents of vendor/
func VerifyVendor() error {
	output, _ := must.OutputE("git", "status", "--porcelain", "vendor")
	if output != "" {
		return fmt.Errorf("vendor directory is out-of-date:\n%s", output)
	}
	return nil
}
