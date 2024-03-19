package logger

import "fmt"

const (
	red        = "\033[31m"
	green      = "\033[32m"
	yellow     = "\033[33m"
	resetColor = "\033[0m"
)

func Red(str string) string {
	return fmt.Sprintf("%s%s%s", red, str, resetColor)
}

func Green(str string) string {
	return fmt.Sprintf("%s%s%s", green, str, resetColor)
}

func Yellow(str string) string {
	return fmt.Sprintf("%s%s%s", yellow, str, resetColor)
}
