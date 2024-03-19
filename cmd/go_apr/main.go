package main

import (
	"flag"

	fl "github.com/99pouria/go-apr/internal/fault_localizer"
	preprocess "github.com/99pouria/go-apr/internal/pre-process"
	env "github.com/99pouria/go-apr/internal/projectenv"
	"github.com/99pouria/go-apr/pkg/logger"
)

func main() {

	// logger.EnableDebugMode()

	fileName := flag.String("p", "", "Path to golang file that contains function for test")
	funcName := flag.String("f", "", "Name of function which needs repair")
	testFile := flag.String("t", "", "Path to test files which contains test cases for given Golang function")

	flag.Parse()

	logger.Printf("Checking go file...\t\t")

	goCode, err := preprocess.StartPreProcess(*fileName, *funcName)
	if err != nil {
		logger.Fatalf("%s\t%s", logger.Symbols.Cross, err)
	}

	logger.Println(logger.Symbols.Tick)

	logger.Printf("Creating build environment...\t")

	be, err := env.CreateEnvironment(*goCode, *testFile)
	if err != nil {
		logger.Fatalf("%s\n\t%s %s", logger.Symbols.Cross, logger.Red("[ERROR]"), err)
	}
	defer be.Finilize(*fileName)

	logger.Println(logger.Symbols.Tick)

	fl.LocalizeFaults(be)
}
