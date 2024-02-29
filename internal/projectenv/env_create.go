package projectenv

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/99pouria/go-apr/internal/code"
)

const (
	envDirPattern string = "apr_go_*"
	moduleName    string = "apr_go_build_env"
)

type Environment struct {
	rootPath string

	FuncCode *code.Code

	TestCases []TestCase
}

type TestCase struct {
	ID      int
	Inputs  []string
	Outputs []string
}

func CreateEnvironment(goCode code.Code, testCasePath string) (*Environment, error) {
	// preparing root dir for build env
	rootDir, err := os.MkdirTemp(os.TempDir(), envDirPattern)
	if err != nil {
		return nil, err
	}

	// creating main function
	fd, err := os.OpenFile(
		filepath.Join(rootDir, "main.go"),
		os.O_CREATE|os.O_WRONLY,
		0755,
	)
	if err != nil {
		return nil, fmt.Errorf("can not create main file: %w", err)
	}
	defer fd.Close()

	if _, err := fd.WriteString(generateMainFunction(goCode)); err != nil {
		return nil, fmt.Errorf("can not write main func content: %w", err)
	}

	// create package directory
	if err := os.Mkdir(filepath.Join(rootDir, goCode.PackageName), 0755); err != nil {
		return nil, fmt.Errorf("can not create dir for given file: %w", err)
	}

	// copy given go file to desired destination
	dfd, err := os.OpenFile(
		filepath.Join(rootDir, goCode.PackageName, filepath.Base(goCode.Path)),
		os.O_CREATE|os.O_WRONLY,
		0755,
	)
	if err != nil {
		return nil, fmt.Errorf("can not create package file: %w", err)
	}
	defer dfd.Close()

	sfd, err := os.Open(goCode.Path)
	if err != nil {
		return nil, fmt.Errorf("can not open src file: %w", err)
	}
	defer sfd.Close()

	if _, err := io.Copy(dfd, sfd); err != nil {
		return nil, fmt.Errorf("can not copy src file content to dst file: %w", err)
	}

	// create new Code object
	destCode, err := code.NewCode(dfd.Name(), goCode.FuncName)
	if err != nil {
		return nil, fmt.Errorf("can not create Code object in build env: %w", err)
	}

	// download required modules
	if err := prepareGoModule(rootDir); err != nil {
		return nil, fmt.Errorf("can not prepare go module: %w", err)
	}

	// parse testcases
	testCases, err := parseTestCases(testCasePath)
	if err != nil {
		return nil, err
	}

	be := new(Environment)

	be.rootPath = rootDir
	be.FuncCode = destCode
	be.TestCases = testCases

	return be, nil
}

func prepareGoModule(envPath string) error {
	command := fmt.Sprintf("cd %s && go mod init %s && go mod tidy", envPath, moduleName)
	out, err := exec.Command("bash", "-c", command).CombinedOutput()
	if err != nil {
		return fmt.Errorf("go mod failed: %s", out)
	}

	return nil
}

func parseTestCases(path string) ([]TestCase, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := bytes.Split(content, []byte("\n"))
	testCases := make([]TestCase, 0)
	for i := 0; i < len(lines); i += 2 {
		testCases = append(testCases, TestCase{
			ID:      (i / 2) + 1,
			Inputs:  strings.Split(string(lines[i]), ","),
			Outputs: strings.Split(string(lines[i+1]), ","),
		})
	}

	return testCases, nil
}
