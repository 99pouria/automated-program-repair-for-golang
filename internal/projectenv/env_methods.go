package projectenv

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/99pouria/go-apr/pkg/logger"
	"github.com/99pouria/go-apr/utils"
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
//   - debugMode true means that test cases doesn't need to check if they are passed
//   - n is number of run for each testcase
func (env *Environment) RunTestCases(debugMode bool, n int) []ExecutionResult {
	var result []ExecutionResult

	if !debugMode {
		logger.Printf("Running testcases %d time(s)...\n", n)
		logger.Println("==============================================")
		logger.Println("TestID\tRound\tOK\t Description")
		logger.Println("----------------------------------------------")
	}

	for _, testCase := range env.TestCases {
		result = append(result, env.RunTestCase(testCase.ID, debugMode, n))
	}
	// TODO: print some percentage status for result
	return result
}

// RunTestCase runs testcase that its ID is given as input of the function
//   - testID is id of testcase
//   - debugMode true means that test cases doesn't need to check if they are passed
//   - n is number of run for each testcase
func (env *Environment) RunTestCase(testID int, debugMode bool, n int) ExecutionResult {

	result := ExecutionResult{
		TestCase: env.TestCases[testID-1],
		Ok:       false,
	}

	if err := utils.FixImports(env.FuncCode.Path); err != nil {
		result.Err = err
		return result
	}

	if err := env.BuildProject(); err != nil {
		result.Err = err
		return result
	}

	execCommand := filepath.Join(env.rootPath, "app")
	for _, input := range env.TestCases[testID-1].Inputs {
		execCommand = fmt.Sprintf("%s %s", execCommand, input)
	}

	result.Ok = true

	// running testcases for n times
	for round := range n {
		if !debugMode {
			logger.Printf("\r%d\t%d\t", testID, round+1)
		}

		out, err := exec.Command("bash", "-c", execCommand).CombinedOutput()
		if err != nil {
			result.Err = err
			return result
		}

		for index, outLine := range strings.Fields(string(out)) {
			if !debugMode && outLine != result.Outputs[index] {
				result.Ok = false
			}
			result.ActualOutputs = append(result.ActualOutputs, outLine)
		}

		if !result.Ok && !debugMode {
			logger.Printf("%s\t%s=%v\t%s=%v\n", logger.Symbols.Cross, logger.Green("Expected"), result.Outputs, logger.Red("Actual"), result.ActualOutputs[len(result.ActualOutputs)-len(result.Outputs):])
			break
		}

	}

	if result.Ok && !debugMode {
		logger.Println(logger.Symbols.Tick)
	}

	return result
}

// Finilize deletes created env and stores repaired file
func (env *Environment) Finilize(path string) error {
	newFileName := fmt.Sprintf("repaired_%s", path)
	fd, err := os.Create(newFileName)
	if err != nil {
		return err
	}

	if _, err := io.Copy(fd, strings.NewReader(env.FuncCode.CodeContent)); err != nil {
		return err
	}

	return os.RemoveAll(env.rootPath)
}
