package utils

import (
	"errors"
	"os"
)

// FileExist returns true if given file exists. Otherwise it returns false.
func FileExist(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}
