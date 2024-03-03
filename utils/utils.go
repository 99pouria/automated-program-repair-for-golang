package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// FileExist returns true if given file exists. Otherwise it returns false.
func FileExist(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

// FormatGoFile uses gofmt command to format given golang file.
func FormatGoFile(path string) error {
	res, err := exec.Command("/bin/bash", "-c", "gofmt", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf(string(res))
	}
	return nil
}

// FixImports uses goimports to fix imports issues
// TODO: move goimports to the repo
func FixImports(path string) error {
	res, err := exec.Command("/Users/pooria/go/bin/goimports", "-w", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", string(res), err)
	}
	return nil
}
