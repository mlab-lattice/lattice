// Package tests implements tests for component resolution.
// It is necessary for it to be its own package due to the
// mock secret/template stores importing the resolver package
// creating a dependency cycle.
package test
