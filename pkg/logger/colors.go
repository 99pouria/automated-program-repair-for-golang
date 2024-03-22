package logger

import "fmt"

const (
	red        = "\033[31m"
	green      = "\033[32m"
	yellow     = "\033[33m"
	blue       = "\033[36m"
	resetColor = "\033[0m"
)

func Red(a any) string {
	return fmt.Sprintf("%s%v%s", red, a, resetColor)
}

func Green(a any) string {
	return fmt.Sprintf("%s%v%s", green, a, resetColor)
}

func Yellow(a any) string {
	return fmt.Sprintf("%s%v%s", yellow, a, resetColor)
}

func Blue(a any) string {
	return fmt.Sprintf("%s%v%s", blue, a, resetColor)
}
