// +build mage

// This is a magefile, and is a "makefile for go".
// See https://magefile.org/
package main

import (
	"fmt"
	
	"github.com/carolynvs/magex/shx"
)

// Any commands executed by "must" (as opposed to shx.RunV for example), will stop
// the build immediately when the command fails.
var must = shx.CommandBuilder{StopOnError: true}

func Vendor() {
	must.RunV("go", "mod", "vendor")
}

// Check if go.mod matches the contents of vendor/
func VerifyVendor() error {
	output, _:= must.OutputE("git", "status", "--porcelain")
	if output != "" {
		return fmt.Errorf("vendor directory is out-of-date:\n%s", output)
	}
	return nil
}
