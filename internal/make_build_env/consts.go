package makebuildenv

import (
	"fmt"
	"os"
)

const (
	moduleName string = "apr_go_build_env"

	// todo: use os.Stderr for checking error of build file
	// or maybe check error codes
	mainFileContent string = `
package main

import (
	"os"

	xxx "%s/%s"
)

func main() {

	if len(os.Args) < %d + 1{
		os.Exit(1)
	}

	for i := 1; i < len(os.Args); i++ {
		argsString += "," + os.Args[i]
	}

	argsString = strings.TrimLeft(argsString, ",")


	// fill inputs comma separed
	fmt.Println(xxx.%s(%s))
}
`
)

func getMainFileContent(packageName, funcName string, args ...interface{}) string {
	return fmt.Sprintf(mainFileContent, moduleName, packageName, len(os.Args), funcName)
}
