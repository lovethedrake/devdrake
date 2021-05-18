// +build mage

// This is a magefile, and is a "makefile for go".
// See https://magefile.org/
package main

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/carolynvs/magex/shx"

	// Import shared mage targets for drake
	// mage:import
	"github.com/lovethedrake/drakecore/mage"
)

// Any commands executed by "must" (as opposed to shx.RunV for example), will stop
// the build immediately when the command fails.
var must = shx.CommandBuilder{StopOnError: true}

// Compile the mallard CLI with Docker
func Build() {
	pwd, _ := os.Getwd()
	must.RunV("docker", "run", "--rm",
		"-v", pwd+":/go/src/github.com/lovethedrake/mallard",
		"-w", "/go/src/github.com/lovethedrake/mallard",
		"-v", pwd+"/bin:/shared/bin",
		"brigadecore/go-tools:v0.1.0", "scripts/build.sh", runtime.GOOS, runtime.GOARCH)
}

// Check go code for lint errors
func Lint() {
	must.RunV("golangci-lint", "run")
}

// Run unit tests
func Test() {
	coverageFile := filepath.Join(mage.GetOutputDir(), "coverage.txt")
	must.RunV("go", "test", "-timeout=30s", "-race", "-coverprofile", coverageFile, "-covermode=atomic", "./cmd/...", "./pkg/...")
}
