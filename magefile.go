// +build mage

// This is a magefile, and is a "makefile for go".
// See https://magefile.org/
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/carolynvs/magex/shx"
)

// Any commands executed by "must" (as opposed to shx.RunV for example), will stop
// the build immediately when the command fails.
var must = shx.CommandBuilder{StopOnError: true}

// Cache dependencies to vendor/
func Vendor() {
	must.RunV("go", "mod", "vendor")
}

// Check if go.mod matches the contents of vendor/
func VerifyVendor() error {
	output, _ := must.OutputE("git", "status", "--porcelain")
	if output != "" {
		return fmt.Errorf("vendor directory is out-of-date:\n%s", output)
	}
	return nil
}

// Compile the drake CLI with Docker
func Build() {
	pwd, _ := os.Getwd()
	must.RunV("docker", "run", "--rm",
		"-v", pwd+":/go/src/github.com/lovethedrake/devdrake",
		"-w", "/go/src/github.com/lovethedrake/devdrake",
		"-v", pwd+"/bin:/shared/bin/drake",
		"brigadecore/go-tools:v0.1.0", "scripts/build.sh", runtime.GOOS, runtime.GOARCH)
}

// Run unit tests
func Test() {
	coverageFile := filepath.Join(getOutputDir(), "coverage.txt")
	must.RunV("go", "test", "-timeout=30s", "-race", "-coverprofile", coverageFile, "-covermode=atomic", "./cmd/...", "./pkg/...")
}

var outDir = ""

func getOutputDir() string {
	if outDir != "" {
		return outDir
	}
	const sharedVolume = "/shared"
	if _, err := os.Stat(sharedVolume); err == nil {
		return sharedVolume
	}

	return "."
}
