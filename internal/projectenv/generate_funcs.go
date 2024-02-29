package projectenv

import (
	"fmt"

	"github.com/99pouria/go-apr/internal/code"
	"github.com/dave/jennifer/jen"
)

// generateMainFunction creates body of a main function that does these operations:
//   - checks if length of args is equal to go function inputs
//   - all args is string at first; these strings converts to specific types that is supported by the apr
//     and also the go function accepts the types. for example, if function gets int64 and bool as input, genrator
//     generates two type conversions that converts first arg from string to int64 and second arg to bool.
//   - call the go function and pass casted args to it.
//   - implement an exit function that writes output to Stdout and exits from the program with given status code.
func generateMainFunction(goCode code.Code) string {
	f := jen.NewFile("main")

	mainFunc := f.Func().Id("main").Params()
	exitFunc := f.Func().Id("exit").Params(jen.Id("statusCode").Int(), jen.Id("out").Op("...").Any())

	var mainFuncBlock []jen.Code

	numberOfInputs := 0
	for _, count := range goCode.InputTypes {
		numberOfInputs += count
	}

	numberOfOutputs := 0
	for _, count := range goCode.OutputTypes {
		if count == 0 {
			count = 1
		}
		numberOfOutputs += count
	}

	checkArgsCode := checkArgsLen(numberOfInputs)
	convertInputsCode := convertInputs(goCode.InputTypes)

	callPkg := callPkgFunc(
		fmt.Sprintf("%s/%s", moduleName, goCode.PackageName),
		goCode.FuncName,
		numberOfInputs,
		numberOfOutputs,
	)

	mainFuncBlock = append(mainFuncBlock, checkArgsCode...)
	mainFuncBlock = append(mainFuncBlock, convertInputsCode...)
	mainFuncBlock = append(mainFuncBlock, callPkg...)

	mainFunc.Block(mainFuncBlock...)
	exitFunc.Block(exit()...)

	return fmt.Sprintf("%#v", f)
}

// exit generates exit function
func exit() []jen.Code {
	defResult := jen.Var().Id("res").String()

	resultFiller := jen.For(jen.List(jen.Id("_"), jen.Id("out_i")).Op(":=").Range().Id("out")).Block(
		jen.Id("res").Op("=").Qual("fmt", "Sprintf").Params(jen.Lit("%s%v\n"), jen.Id("res"), jen.Id("out_i")),
	)

	writeResult := jen.Qual("fmt", "Fprint").Params(
		jen.Qual("os", "Stdout"),
		jen.Id("res"),
	)

	osExitCaller := jen.Qual("os", "Exit").Params(jen.Id("statusCode"))

	return []jen.Code{defResult, resultFiller, writeResult, osExitCaller}
}

// callPkgFunc generates a code that calls the golang function with arguments
func callPkgFunc(pkgName, funcName string, inputLen, outputLen int) []jen.Code {
	params := make([]jen.Code, inputLen)
	for i := 0; i < inputLen; i++ {
		params[i] = jen.Id(fmt.Sprintf("in%d", i+1))
	}

	outs := make([]jen.Code, outputLen)
	for i := 0; i < outputLen; i++ {
		outs[i] = jen.Id(fmt.Sprintf("out%d", i+1))
	}

	funcCaller := jen.List(outs...).Op(":=").Qual(pkgName, funcName).Params(params...)

	outs = append([]jen.Code{jen.Lit(0)}, outs...)
	returnReuslt := jen.Id("exit").Call(outs...)

	return []jen.Code{funcCaller, returnReuslt}
}

// checkArgsLen generates input length checker
func checkArgsLen(inputLen int) []jen.Code {
	ifStatement := jen.If(jen.Len(jen.Qual("os", "Args")).Op("!=").Lit(inputLen + 1)).Block(
		jen.Id("exit").Call(jen.Lit(2), jen.Lit("not enough arguments")),
	)

	return []jen.Code{ifStatement}
}

// convertInputs generates a code that converts all inputs to supported types
func convertInputs(inputs map[string]int) []jen.Code {
	inputParserCode := make([]jen.Code, 0, len(inputs))
	inputIndex := 0
	for itype, count := range inputs {
		for range count {
			inputIndex += 1

			inputName := fmt.Sprintf("in%d", inputIndex)
			strInput := fmt.Sprintf("os.Args[%d]", inputIndex)
			var convertCode []jen.Code

			switch itype {
			case "bool":
				convertCode = str2Bool(inputName, strInput)
			case "string":
				convertCode = str2str(inputName, strInput)
			case "int":
				convertCode = str2intBitSize(inputName, strInput, 0)
			case "int8":
				convertCode = str2intBitSize(inputName, strInput, 8)
			case "int16":
				convertCode = str2intBitSize(inputName, strInput, 16)
			case "int32":
				convertCode = str2intBitSize(inputName, strInput, 32)
			case "int64":
				convertCode = str2intBitSize(inputName, strInput, 64)
			case "uint":
				convertCode = str2UintBitSize(inputName, strInput, 0)
			case "uint8":
				convertCode = str2UintBitSize(inputName, strInput, 8)
			case "uint16":
				convertCode = str2UintBitSize(inputName, strInput, 16)
			case "uint32":
				convertCode = str2UintBitSize(inputName, strInput, 32)
			case "uint64":
				convertCode = str2UintBitSize(inputName, strInput, 64)
			case "float32":
				convertCode = str2floatBitSize(inputName, strInput, 32)
			case "float64":
				convertCode = str2floatBitSize(inputName, strInput, 64)
			}

			comment := jen.Comment(fmt.Sprintf("converting input %d to %s", inputIndex, itype))

			inputParserCode = append(inputParserCode, comment)
			inputParserCode = append(inputParserCode, convertCode...)
		}
	}

	return inputParserCode
}

/*

	All functions bellow generates a code that converts string to one of these types:
	int, int8, uint16, int32, int64
	uint, uint8, uint16, uint32, uint64
	float32, float64
	bool
	string

*/

func str2str(resName string, str string) []jen.Code {
	assigner := jen.Id(resName).Op(":=").Id(str)

	return []jen.Code{assigner}
}

func str2Bool(resName string, str string) []jen.Code {
	caller := jen.List(jen.Id(resName), jen.Err()).Op(":=").Qual("strconv", "ParseBool").Params(jen.Id(str))
	errHandler := handleError()

	return []jen.Code{caller, errHandler}
}

func str2floatBitSize(resName string, str string, bitSize int) []jen.Code {
	var varDefiner jen.Code
	switch bitSize {
	case 32:
		varDefiner = jen.Var().Id(resName).Float32()
	case 64:
		varDefiner = jen.Var().Id(resName).Float64()
	}
	unConverted := fmt.Sprintf("u_%s", resName)

	caller := jen.List(jen.Id(unConverted), jen.Err()).Op(":=").Qual("strconv", "ParseFloat").Params(
		jen.Id(str),
		jen.Lit(bitSize), // Bit size
	)

	var assigner jen.Code
	switch bitSize {
	case 32:
		assigner = jen.Id(resName).Op("=").Float32().Parens(jen.Id(unConverted))
	case 64:
		assigner = jen.Id(resName).Op("=").Float64().Parens(jen.Id(unConverted))
	}

	errHandler := handleError()

	return []jen.Code{varDefiner, caller, errHandler, assigner}
}

func str2intBitSize(resName string, str string, bitSize int) []jen.Code {
	var varDefiner jen.Code
	switch bitSize {
	case 8:
		varDefiner = jen.Var().Id(resName).Int8()
	case 16:
		varDefiner = jen.Var().Id(resName).Int16()
	case 32:
		varDefiner = jen.Var().Id(resName).Int32()
	case 64:
		varDefiner = jen.Var().Id(resName).Int64()
	default:
		varDefiner = jen.Var().Id(resName).Int()
	}

	unConverted := fmt.Sprintf("u_%s", resName)

	caller := jen.List(jen.Id(unConverted), jen.Err()).Op(":=").Qual("strconv", "ParseInt").Params(
		jen.Id(str),
		jen.Lit(10),      // Base  10
		jen.Lit(bitSize), // Bit size
	)
	var assigner jen.Code
	switch bitSize {
	case 8:
		assigner = jen.Id(resName).Op("=").Int8().Parens(jen.Id(unConverted))
	case 16:
		assigner = jen.Id(resName).Op("=").Int16().Parens(jen.Id(unConverted))
	case 32:
		assigner = jen.Id(resName).Op("=").Int32().Parens(jen.Id(unConverted))
	case 64:
		assigner = jen.Id(resName).Op("=").Int64().Parens(jen.Id(unConverted))
	default:
		assigner = jen.Id(resName).Op("=").Int().Parens(jen.Id(unConverted))
	}

	errHandler := handleError()

	return []jen.Code{varDefiner, caller, errHandler, assigner}
}

func str2UintBitSize(resName string, str string, bitSize int) []jen.Code {
	var varDefiner jen.Code
	switch bitSize {
	case 8:
		varDefiner = jen.Var().Id(resName).Uint8()
	case 16:
		varDefiner = jen.Var().Id(resName).Uint16()
	case 32:
		varDefiner = jen.Var().Id(resName).Uint32()
	case 64:
		varDefiner = jen.Var().Id(resName).Uint64()
	default:
		varDefiner = jen.Var().Id(resName).Uint()
	}

	unConverted := fmt.Sprintf("u_%s", resName)

	caller := jen.List(jen.Id(unConverted), jen.Err()).Op(":=").Qual("strconv", "ParseUint").Params(
		jen.Id(str),
		jen.Lit(10),      // Base  10
		jen.Lit(bitSize), // Bit size
	)
	var assigner jen.Code
	switch bitSize {
	case 8:
		assigner = jen.Id(resName).Op("=").Uint8().Parens(jen.Id(unConverted))
	case 16:
		assigner = jen.Id(resName).Op("=").Uint16().Parens(jen.Id(unConverted))
	case 32:
		assigner = jen.Id(resName).Op("=").Uint32().Parens(jen.Id(unConverted))
	case 64:
		assigner = jen.Id(resName).Op("=").Uint64().Parens(jen.Id(unConverted))
	default:
		assigner = jen.Id(resName).Op("=").Uint().Parens(jen.Id(unConverted))
	}

	errHandler := handleError()

	return []jen.Code{varDefiner, caller, errHandler, assigner}
}

func handleError() jen.Code {
	return jen.If(jen.Err().Op("!=").Nil()).Block(
		jen.Id("exit").Call(jen.Lit(1), jen.Id("err")),
	)
}
