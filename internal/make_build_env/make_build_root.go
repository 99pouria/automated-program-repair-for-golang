package makebuildenv

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/99pouria/go-apr/internal/code"
)

const (
	envDirPattern string = "apr_go_*"
	moduleName    string = "apr_go_build_env"
)

type BuildEnvironment struct {
	rootPath string

	FuncCode code.Code
}

func NewBuildEnvironment(path, funcName string, goCode code.Code) (*BuildEnvironment, error) {

	// call MakeBuildEnv without passing code.Code then fill the struct and return it

	return nil, nil
}

func (env *BuildEnvironment) Run() {

}

// Destruct deletes created env
func (env *BuildEnvironment) Destruct() error {
	return os.RemoveAll(env.rootPath)
}

func MakeBuildEnv(goCode code.Code) error {
	// preparing root dir for build env

	rootDir := "/Users/pooria/Desktop/apr_go"
	err := os.Mkdir(rootDir, 0755)
	// rootDir, err := os.MkdirTemp("/Users/pooria/Desktop/", envDirPattern)
	// rootDir, err := os.MkdirTemp(os.TempDir(), envDirPattern)
	if err != nil {
		return err
	}

	// creating main function
	fd, err := os.OpenFile(
		filepath.Join(rootDir, "main.go"),
		os.O_CREATE|os.O_WRONLY,
		0755,
	)
	if err != nil {
		return fmt.Errorf("can not create main file: %w", err)
	}
	defer fd.Close()

	if _, err := fd.WriteString(generateMainFunction(goCode)); err != nil {
		return fmt.Errorf("can not write main func content: %w", err)
	}

	// copy given file to desired destination
	// todo: handle main package
	if err := os.Mkdir(filepath.Join(rootDir, goCode.PackageName), 0755); err != nil {
		return fmt.Errorf("can not create dir for given file: %w", err)
	}

	dfd, err := os.OpenFile(
		filepath.Join(rootDir, goCode.Path),
		os.O_CREATE|os.O_WRONLY,
		0755,
	)
	if err != nil {
		return fmt.Errorf("can not create package file: %w", err)
	}
	defer dfd.Close()

	sfd, err := os.Open(goCode.Path)
	if err != nil {
		return fmt.Errorf("can not open src file: %w", err)
	}
	defer sfd.Close()

	if _, err := io.Copy(dfd, sfd); err != nil {
		return fmt.Errorf("can not copy src file content to dst file: %w", err)
	}

	// download required modules
	if err := prepareGoModule(rootDir); err != nil {
		return fmt.Errorf("can not prepare go module: %w", err)
	}

	return nil
}

func prepareGoModule(envPath string) error {
	command := fmt.Sprintf("cd %s && go mod init %s && go mod tidy", envPath, moduleName)
	out, err := exec.Command("bash", "-c", command).CombinedOutput()
	if err != nil {
		return fmt.Errorf("go mod failed: %s", out)
	}

	return nil
}
