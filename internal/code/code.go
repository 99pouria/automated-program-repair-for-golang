package code

import (
	"fmt"
	"go/types"
	"os"
	"strings"

	preprocess "github.com/99pouria/go-apr/internal/pre-process"
)

type Code struct {
	Path        string
	FuncName    string
	CodeContent string

	StartOfFuncLine int
	EndOfFuncLine   int

	InputTypes  []types.Type
	OutputTypes []types.Type
}

func NewCode(path, funcName string) (*Code, error) {
	c := new(Code)

	c.Path = path
	c.FuncName = funcName

	if err := c.updateFuncPosition(); err != nil {
		return nil, err
	}

	if err := c.findFuncIOType(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Code) findFuncIOType() error {
	if err := c.updateFuncPosition(); err != nil {
		return err
	}

	// finding inputs
	

	return nil
}

func (c *Code) updateCodeContentFromPath() error {
	if err := preprocess.FormatGoFile(c.Path); err != nil {
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

	return preprocess.FormatGoFile(c.Path)
}
