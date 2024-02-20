package makebuildenv

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

const (
	envDirPattern string = "apr_go_*"
)

type BuildEnvironment struct {
	rootPath string
}

func NewBuildEnvironment(path, funcName string) (*BuildEnvironment, error) {

	return nil, nil
}

// Destruct
func (env *BuildEnvironment) Destruct() error {
	return os.RemoveAll(env.rootPath)
}

func MakeBuildEnv(fileName, funcName string, args ...interface{}) error {
	// finding package name
	packageName, err := findPackageName(fileName)
	if err != nil {
		return err
	}

	// preparing root dir for build env
	rootDir, err := os.MkdirTemp("", envDirPattern)
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

	if _, err := fd.WriteString(getMainFileContent(packageName, funcName, args)); err != nil {
		return fmt.Errorf("can not write main func content: %w", err)
	}

	// copy given file to desired destination
	// todo: handle main package
	if err := os.Mkdir(filepath.Join(rootDir, packageName), 0755); err != nil { // todo: perm must be constant?
		return fmt.Errorf("can not create dir for given file: %w", err)
	}

	dfd, err := os.OpenFile(
		filepath.Join(rootDir, packageName, fileName),
		os.O_CREATE|os.O_WRONLY,
		0755,
	)
	if err != nil {
		return fmt.Errorf("can not create package file: %w", err)
	}
	defer dfd.Close()

	sfd, err := os.Open(fileName)
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

func findPackageName(fileName string) (string, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}

	re, err := regexp.Compile(`^\s*package\s+(\w+)`)
	if err != nil {
		return "", err
	}

	pkgName := re.FindStringSubmatch(string(data))

	if len(pkgName) < 1 {
		return "", fmt.Errorf("can not find pacakge name")
	}

	return pkgName[0], nil
}
