package main

import (
	"flag"
	"fmt"

	preprocess "github.com/99pouria/go-apr/internal/pre-process"
	env "github.com/99pouria/go-apr/internal/projectenv"
	"github.com/sirupsen/logrus"
)

func main() {
	fileName := flag.String("p", "", "Path to golang file that contains function for test")
	funcName := flag.String("f", "", "Name of function which needs repair")
	testFile := flag.String("t", "", "Path to test files which contains test cases for given Golang function")

	flag.Parse()

	fmt.Printf("Checking go file...\t\t")

	goCode, err := preprocess.StartPreProcess(*fileName, *funcName)
	if err != nil {
		logrus.WithField("error", err).Fatal("pre process failed")
	}

	fmt.Printf("OK\n")
	fmt.Printf("Creating build environment...\t")

	be, err := env.CreateEnvironment(*goCode, *testFile)
	if err != nil {
		logrus.WithField("error", err).Fatal("make build env failed")
	}
	defer be.Destruct()

	fmt.Printf("OK\n")
}
