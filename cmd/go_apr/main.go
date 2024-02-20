package main

import (
	"flag"
	"fmt"

	makebuildenv "github.com/99pouria/go-apr/internal/make_build_env"
)

func main() {
	fileName := flag.String("p", "", "Path to golang file that contains function for test")
	funcName := flag.String("f", "", "Name of function which needs repair")
	testFile := flag.String("t", "", "Path to test files which contains test cases for given Golang function")
	_ = testFile
	flag.Parse()

	if err := makebuildenv.MakeBuildEnv(*fileName, *funcName); err != nil {
		fmt.Println(err)
		return
	}

	// strconv.ParseInt()
}
