// TODO: add comment for funcs
package logger

import (
	"fmt"
	"io"
	"os"
)

var (
	ioWriter  io.Writer
	debugMode bool
)

func init() {
	debugMode = false
	ioWriter = os.Stdout
}

func EnableDebugMode() {
	debugMode = true
}

func DisableDebugMode() {
	debugMode = false
}

func AddOutputs(ioOuts ...io.Writer) {
	ioWriter = io.MultiWriter(append(ioOuts, ioWriter)...)
}

func Printf(format string, a ...any) (n int, err error) {
	return fmt.Fprintf(ioWriter, format, a...)
}

func Println(a ...any) (n int, err error) {
	return fmt.Fprintln(ioWriter, a...)
}

func Fatalf(format string, a ...any) {
	fmt.Fprintf(ioWriter, format, a...)
	fmt.Fprintf(ioWriter, "\n")
	os.Exit(1)
}

func Fatal(a ...any) {
	fmt.Fprintln(ioWriter, a...)
	os.Exit(1)
}

func Debugf(format string, a ...any) (n int, err error) {
	if !debugMode {
		return 0, fmt.Errorf("debug mode is disabled")
	}

	fmt.Fprintf(ioWriter, "[%s] ", Yellow("DEBUG"))
	return fmt.Fprintf(ioWriter, format, a...)
}

func Debugln(a ...any) (n int, err error) {
	if !debugMode {
		return 0, fmt.Errorf("debug mode is disabled")
	}

	fmt.Fprintf(ioWriter, "[%s] ", Yellow("DEBUG"))
	return fmt.Fprintln(ioWriter, a...)
}

func Warnf(format string, a ...any) (n int, err error) {
	return fmt.Fprintf(ioWriter, fmt.Sprintf("%s %s", Yellow("[WARN]"), format), a...)
}

func Warnln(a ...any) (n int, err error) {
	return fmt.Fprintln(ioWriter, append([]any{Yellow("[WARN]")}, a...)...)
}
