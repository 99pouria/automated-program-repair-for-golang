package projectenv

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ExecutionResult contains result of project execution
type ExecutionResult struct {
	TestCase
	Ok            bool
	ActualOutputs []string
	Err           error
}

// BuildProject build project using 'go build' command
func (env *Environment) BuildProject() error {
	command := fmt.Sprintf("cd %s && go build -o app main.go", env.rootPath)
	out, err := exec.Command("bash", "-c", command).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", out)
	}

	return nil
}

// RunTestCases runs all testcases of project and returns a slice of ExecutionResult
func (env *Environment) RunTestCases() []ExecutionResult {
	var result []ExecutionResult

	for _, testCase := range env.TestCases {
		result = append(result, env.RunTestCase(testCase.ID))
	}

	return result
}

// RunTestCase runs testcase that its ID is given as input of the function
func (env *Environment) RunTestCase(testID int) ExecutionResult {
	result := ExecutionResult{
		TestCase: env.TestCases[testID-1],
		Ok:       false,
	}

	if err := env.BuildProject(); err != nil {
		result.Err = err
		return result
	}

	execCommand := filepath.Join(env.rootPath, "app")
	for _, input := range env.TestCases[testID-1].Inputs {
		execCommand = fmt.Sprintf("%s %s", execCommand, input)
	}

	out, err := exec.Command("bash", "-c", execCommand).CombinedOutput()
	if err != nil {
		result.Err = err
		return result
	}

	result.Ok = true

	for index, outLine := range strings.Fields(string(out)) {
		if outLine != result.Outputs[index] {
			result.Ok = false
		}
		result.ActualOutputs = append(result.ActualOutputs, outLine)
	}

	return result
}

// Destruct deletes created env
func (env *Environment) Destruct() error {
	return os.RemoveAll(env.rootPath)
}
