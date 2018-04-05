// +build mage

// nolint
package main

import (
	"github.com/aporeto-inc/addedeffect/magetask"
	"github.com/magefile/mage/mg"
)

// Version writes the versions file.
func Version() error {
	return magetask.WriteVersion()
}

// Test runs the unit tests.
func Test() {
	mg.Deps(
		magetask.Lint,
		magetask.Test,
	)
}

// Build builds the project and prepare the docker container.
func Build() error {
	return magetask.Build()
}

// Package prepares the docker container.
func Package() {
	mg.Deps(
		magetask.Package,
		magetask.PackageCACerts,
	)
}

// Docker builds the docker container.
func Docker() error {
	mg.SerialDeps(
		magetask.BuildLinux,
		Package,
	)

	return magetask.Container()
}
