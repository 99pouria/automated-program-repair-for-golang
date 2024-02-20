package preprocess

import (
	"fmt"
	"os/exec"
)

// StartPreProcess does some checks before starting to repair program
func StartPreProcess(codePath, funcName, testCasesPath string) error {
	if err := FormatGoFile(codePath); err != nil {
		return err
	}

	if err := CheckInputs(codePath, funcName, testCasesPath); err != nil {
		return err
	}

	return nil
}

func CheckInputs(codePath, funcName, testCasesPath string) error {

	return nil
}

func FormatGoFile(path string) error {
	res, err := exec.Command("bash", "-c", "gofmt", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf(string(res))
	}
	return nil
}
