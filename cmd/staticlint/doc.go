// staticlint implements multiple needed static checks.
// All checks from golang.org/x/tools/go/analysis/passes
// All SA checks from https://staticcheck.io/docs/checks/
// Check bytes buffer conversions via https://staticcheck.io/docs/checks/#S1030
// Check wrapping errors https://github.com/fatih/errwrap
// Check for database query in loops https://github.com/masibw/goone
// Check for calling os.Exit in main func of main package
// Run command:
// staticlint ./..

package main
