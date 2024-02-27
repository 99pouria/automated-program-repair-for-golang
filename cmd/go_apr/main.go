package main

import (
	"flag"
	"fmt"

	makebuildenv "github.com/99pouria/go-apr/internal/make_build_env"
	preprocess "github.com/99pouria/go-apr/internal/pre-process"
	"github.com/sirupsen/logrus"
)

func main() {
	fileName := flag.String("p", "", "Path to golang file that contains function for test")
	funcName := flag.String("f", "", "Name of function which needs repair")
	testFile := flag.String("t", "", "Path to test files which contains test cases for given Golang function")

	flag.Parse()

	fmt.Printf("Checking go file...\t\t")

	goCode, err := preprocess.StartPreProcess(*fileName, *funcName, *testFile)
	if err != nil {
		logrus.WithField("error", err).Fatal("pre process failed")
	}

	fmt.Printf("OK\n")
	fmt.Printf("Creating build environment...\t")

	if err := makebuildenv.MakeBuildEnv(*goCode); err != nil {
		logrus.WithField("error", err).Fatal("make build env failed")
	}

	fmt.Printf("OK\n")
}
