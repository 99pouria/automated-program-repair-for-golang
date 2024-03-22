package main

import (
	"flag"
	"os"

	fl "github.com/99pouria/go-apr/internal/fault_localizer"
	preprocess "github.com/99pouria/go-apr/internal/pre-process"
	env "github.com/99pouria/go-apr/internal/projectenv"
	"github.com/99pouria/go-apr/pkg/logger"
)

func main() {

	fileName := flag.String("p", "", "Path to golang file that contains function for test")
	funcName := flag.String("f", "", "Name of function which needs repair")
	testFile := flag.String("t", "", "Path to test files which contains test cases for given Golang function")
	save := flag.Bool("save", false, "Saves to source file. By default it only prints repaired code")
	debug := flag.Bool("debug", false, "Enables debug mode to print additional information")

	flag.Parse()

	if *fileName == "" || *funcName == "" || *testFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *debug {
		logger.EnableDebugMode()
	}

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
	defer be.Finilize(*fileName, *debug)

	logger.Println(logger.Symbols.Tick)

	if fl.LocalizeFaults(be) {
		if *save {
			be.FuncCode.SaveToFile(*fileName)
		} else {
			logger.Println("BUG-FREE code:")
			logger.PrintInBoxLeft(be.FuncCode.CodeContent)
		}
	} else if *debug {
		logger.Println("Last code content:")
		logger.PrintInBoxLeft(be.FuncCode.CodeContent)
	}

}
