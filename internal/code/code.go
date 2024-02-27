package code

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/99pouria/go-apr/utils"
	"github.com/sirupsen/logrus"
)

type Code struct {
	Path     string
	FuncName string

	CodeContent string
	PackageName string

	StartOfFuncLine int
	EndOfFuncLine   int

	InputTypes  map[string]int
	OutputTypes map[string]int
}

func NewCode(path, funcName string) (*Code, error) {
	c := new(Code)

	c.Path = path
	c.FuncName = funcName
	c.InputTypes = make(map[string]int)
	c.OutputTypes = make(map[string]int)

	if err := c.updateFuncPosition(); err != nil {
		return nil, err
	}

	if err := c.retrieveTypes(); err != nil {
		_ = err
		return nil, err
	}

	if err := c.retrievePkgName(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Code) retrievePkgName() error {
	re, err := regexp.Compile(`^\s*package\s+([^\d]\w+)`)
	if err != nil {
		return fmt.Errorf("can not compile regex to retrieve package name")
	}

	matches := re.FindStringSubmatch(c.CodeContent)

	switch len(matches) {
	case 0, 1:
		return fmt.Errorf("package name not found")
	case 2:
		c.PackageName = matches[1]
	default:
		c.PackageName = matches[1]
		logrus.WithField("found names", matches).Warn("more than one name for package found")
	}

	return nil
}

func (c *Code) updateCodeContentFromPath() error {
	if err := utils.FormatGoFile(c.Path); err != nil {
		return err
	}

	newContent, err := os.ReadFile(c.Path)
	if err != nil {
		return err
	}
	c.CodeContent = string(newContent)

	return nil
}

func (c *Code) updateFuncPosition() error {

	if err := c.updateCodeContentFromPath(); err != nil {
		return err
	}

	newStart, newEnd := -1, -1

	codeLines := strings.Split(c.CodeContent, "\n")
	for index, line := range codeLines {
		if strings.Contains(line, fmt.Sprintf("func %s(", c.FuncName)) {
			newStart = index + 1
			break
		}
	}

	if newStart == -1 {
		return fmt.Errorf("can not find function")
	}

	for i := newStart; i < len(codeLines); i++ {
		if len(codeLines[i]) > 0 && codeLines[i][0] == '}' {
			newEnd = i + 1
			break
		}
	}

	if newEnd == -1 {
		return fmt.Errorf("can not locate end of function")
	}

	c.StartOfFuncLine, c.EndOfFuncLine = newStart, newEnd

	return nil
}

func (c *Code) ReplaceFuncBody(newBody string) error {

	if err := c.updateFuncPosition(); err != nil {
		return err
	}

	lines := strings.Split(c.CodeContent, "\n")

	c.CodeContent = strings.Join(lines[0:c.StartOfFuncLine], "\n") + "\n" + newBody + "\n" + strings.Join(lines[c.EndOfFuncLine-1:], "\n")
	c.EndOfFuncLine = c.StartOfFuncLine + len(strings.Split(newBody, "\n"))

	if err := os.Truncate(c.Path, 0); err != nil {
		return err
	}

	if err := os.WriteFile(c.Path, []byte(c.CodeContent), 0755); err != nil {
		return err
	}

	return utils.FormatGoFile(c.Path)
}
