package faults

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"math/rand"
	"os"
	"regexp"
	"strings"

	"github.com/99pouria/go-apr/internal/projectenv"
	"github.com/99pouria/go-apr/utils"
	"github.com/sirupsen/logrus"
)

const emptyPattern string = "*_apr_goroutine_done_"

type WG struct {
	env     *projectenv.Environment
	pattern string

	oldFileContent string

	goroutineCount int

	re *regexp.Regexp
}

func InitWaitGroupFault(env *projectenv.Environment) *WG {
	wg := new(WG)

	wg.env = env
	wg.pattern = emptyPattern
	wg.oldFileContent = env.FuncCode.CodeContent

	for ; strings.Contains(env.FuncCode.CodeContent, wg.pattern); wg.pattern = strings.Replace(emptyPattern, "*", fmt.Sprint(rand.Intn(1000)), 1) {
	}

	wg.re = regexp.MustCompile(fmt.Sprintf(`/%s(\d+)`, wg.pattern))

	return wg
}

func (wg *WG) Check() (bool, error) {
	logrus.WithField("fault", "wait-group").Info("Looking for unfinished goroutines...")

	// Create the AST by parsing the source
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", wg.env.FuncCode.CodeContent, 0)
	if err != nil {
		return false, err
	}

	// finding position of target function
	var start, end token.Pos
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Name.Name == wg.env.FuncCode.FuncName {
				start, end = fn.Pos(), fn.End()
			}
		}
		return true
	})

	// Initialize goroutine ID
	goroutineID := 1

	// Find and modify goroutines in target function
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncLit); ok && fn.Pos() >= start && fn.End() <= end {
			// Found a goroutine
			logrus.WithFields(logrus.Fields{
				"fault":              "wait-group",
				"goroutine position": fset.Position(fn.Pos()),
			}).Debug("Found a goroutine")

			// Add printf("apr_goroutine_%d_done\n", id) at the beginning of goroutine
			printfCall := &ast.ExprStmt{
				X: &ast.CallExpr{
					Fun:  &ast.Ident{Name: "defer println"},
					Args: []ast.Expr{&ast.BasicLit{Value: fmt.Sprintf("\"%s%d\"", wg.pattern, goroutineID)}},
				},
			}
			fn.Body.List = append([]ast.Stmt{printfCall}, fn.Body.List...)

			// Increase goroutine ID
			goroutineID++
		}
		return true
	})

	wg.goroutineCount = goroutineID - 1

	// Store changes
	fd, err := os.Create(wg.env.FuncCode.Path)
	if err != nil {
		return false, err
	}
	defer fd.Close()

	// Format and write the modified AST
	if err := format.Node(fd, fset, file); err != nil {
		return false, err
	}

	logrus.WithField("fault", "wait-group").Debug("Successfully modified the file")

	defer wg.Revert()

	results := wg.env.RunTestCases(true, 1)
	for _, result := range results {
		rawResult := strings.Join(result.ActualOutputs, " ")
		matches := wg.re.FindAllStringSubmatch(rawResult, -1)
		x := make(map[string]bool)
		for _, match := range matches {
			if len(match) > 1 {
				x[match[1]] = true
			}
		}

		for id := range wg.goroutineCount {
			if !x[fmt.Sprint(id+1)] {
				logrus.WithFields(logrus.Fields{
					"fault":           "wait-group",
					"goroutine ID":    id + 1,
					"testcase number": result.ID,
				}).Warn("A goroutine doesn't finished completely")
				return false, nil
			}
		}
	}

	logrus.WithField("fault", "wait-group").Info("All goroutines worked completely")

	return true, nil
}

func (wg *WG) Fix() error {
	if err := wg.Revert(); err != nil {
		return err
	}
	// Create the AST by parsing the source
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", wg.env.FuncCode.CodeContent, 0)
	if err != nil {
		return fmt.Errorf("error parsing file: %v", err)
	}

	// finding position of target function
	var start, end token.Pos
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Name.Name == wg.env.FuncCode.FuncName {
				start, end = fn.Pos(), fn.End()
			}
		}
		return true
	})

	// Add sync.WaitGroup variable at the beginning of target function
	ast.Inspect(file, func(n ast.Node) bool {
		// Add var wg_apr sync.WaitGroup and wg_apr.Add(2) at the beginning of the function
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Pos() >= start && fn.End() <= end {
			stmt := &ast.DeclStmt{
				Decl: &ast.GenDecl{
					Tok:   token.VAR,
					Specs: []ast.Spec{&ast.ValueSpec{Names: []*ast.Ident{ast.NewIdent("wg_apr")}, Type: ast.NewIdent("sync.WaitGroup")}},
				},
			}
			fn.Body.List = append([]ast.Stmt{stmt, &ast.ExprStmt{
				X: &ast.CallExpr{
					Fun:  &ast.SelectorExpr{X: ast.NewIdent("wg_apr"), Sel: ast.NewIdent("Add")},
					Args: []ast.Expr{&ast.BasicLit{Value: fmt.Sprint(wg.goroutineCount)}},
				},
			}}, fn.Body.List...)
		}
		return true
	})

	// Add defer wg.Done on each goroutine
	ast.Inspect(file, func(n ast.Node) bool {
		if goStmt, ok := n.(*ast.FuncLit); ok && goStmt.Pos() >= start && goStmt.End() <= end {
			doneCaller := &ast.ExprStmt{X: &ast.CallExpr{Fun: &ast.Ident{Name: "defer wg_apr.Done"}}}
			goStmt.Body.List = append([]ast.Stmt{doneCaller}, goStmt.Body.List...)
		}
		return true
	})

	// Add wg_apr.Wait() before the return statement of target function
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Name.Name == wg.env.FuncCode.FuncName {
				waitCaller := &ast.ExprStmt{X: &ast.CallExpr{Fun: &ast.SelectorExpr{X: ast.NewIdent("wg_apr"), Sel: ast.NewIdent("Wait")}}}

				blockStmtLen := len(fn.Body.List)
				if blockStmtLen != 0 {
					if a, ok := fn.Body.List[blockStmtLen-1].(*ast.ReturnStmt); ok {
						fn.Body.List = append(fn.Body.List[:blockStmtLen-1], waitCaller, a)
					}
				}

			}
		}
		return true
	})

	// Write the modified AST back to the file
	fd, err := os.Create(wg.env.FuncCode.Path)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer fd.Close()

	// Format and write the modified AST
	if err := format.Node(fd, fset, file); err != nil {
		return fmt.Errorf("error writing modified AST: %v", err)
	}

	if err := utils.FixImports(wg.env.FuncCode.Path); err != nil {
		return fmt.Errorf("can not fix imoprts of %s: %w", wg.env.FuncCode.Path, err)
	}

	if err := wg.env.FuncCode.UpdateCodeContentFromPath(); err != nil {
		return fmt.Errorf("can not update code content: %w", err)
	}

	return nil
}

func (wg *WG) Description() string {
	return "unfinished goroutines"
}

func (wg *WG) Revert() error {
	fd, err := os.Create(wg.env.FuncCode.Path)
	if err != nil {
		return err
	}
	if _, err := fd.Write([]byte(wg.oldFileContent)); err != nil {
		return err
	}

	if err := wg.env.FuncCode.UpdateCodeContentFromPath(); err != nil {
		return fmt.Errorf("can not update code content: %w", err)
	}
	return nil
}
