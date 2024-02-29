package preprocess

import (
	"fmt"

	"github.com/99pouria/go-apr/internal/code"
	"github.com/99pouria/go-apr/utils"
)

var SupportedTypes map[string]bool = map[string]bool{
	"string":  true,
	"int":     true,
	"int8":    true,
	"int16":   true,
	"int32":   true,
	"int64":   true,
	"uint":    true,
	"uint8":   true,
	"uint16":  true,
	"uint32":  true,
	"uint64":  true,
	"float32": true,
	"float64": true,
	"bool":    true,
}

// StartPreProcess does some checks before starting to repair program
func StartPreProcess(codePath, funcName string) (*code.Code, error) {
	if err := utils.FormatGoFile(codePath); err != nil {
		return nil, err
	}

	c, err := code.NewCode(codePath, funcName)
	if err != nil {
		return nil, err
	}

	for inputType := range c.InputTypes {
		if !SupportedTypes[inputType] {
			return nil, fmt.Errorf("unsupported type in function input argument: %s", inputType)
		}
	}

	for outputType := range c.OutputTypes {
		if !SupportedTypes[outputType] {
			return nil, fmt.Errorf("unsupported type in function input argument: %s", outputType)
		}
	}

	return c, nil
}
