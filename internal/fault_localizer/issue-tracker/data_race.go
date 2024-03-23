package issuetracker

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/99pouria/go-apr/internal/projectenv"
	"github.com/99pouria/go-apr/utils"
)

type DataRace struct {
	env             *projectenv.Environment
	oldFileContent  string
	containingLines []int

	raceRe *regexp.Regexp
	varRe  *regexp.Regexp
}

func InitDataRaceIT(env *projectenv.Environment) *DataRace {
	return &DataRace{
		env:            env,
		oldFileContent: env.FuncCode.CodeContent,
		raceRe:         regexp.MustCompile(`((/[\w\-\.]+)+):(\d+)`),
		varRe:          regexp.MustCompile(`(var\s+%s|%s.*:=)`),
	}
}

func (d *DataRace) Description() string {
	return "data race"
}

func (d *DataRace) Check() (bool, error) {
	d.oldFileContent = d.env.FuncCode.CodeContent
	// running one of testcases with -race flag
	command := fmt.Sprintf("cd %s && go run -race main.go %s", d.env.RootPath, strings.Join(d.env.TestCases[0].Inputs, " "))
	cmd := exec.Command("bash", "-c", command)
	out, _ := cmd.CombinedOutput()

	if !strings.Contains(string(out), "WARNING: DATA RACE") {
		return true, nil
	}

	res := d.raceRe.FindAllStringSubmatch(string(out), -1)
	for _, r := range res {
		if !strings.Contains(r[2], "main.go") {
			i, _ := strconv.Atoi(r[3])
			d.containingLines = append(d.containingLines, i)
		}
	}
	return false, nil
}

func (d *DataRace) Fix() error {
	d.oldFileContent = d.env.FuncCode.CodeContent
	var lines [][]string
	for index, line := range strings.Split(d.env.FuncCode.CodeContent, "\n") {
		if slices.Contains(d.containingLines, index+1) {
			lines = append(lines, parse(line))
		}
	}

	common, err := FindIntersection(lines)
	if err != nil {
		return err
	}

	// Create the AST by parsing the source
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", d.env.FuncCode.CodeContent, 0)
	if err != nil {
		return fmt.Errorf("error parsing file: %v", err)
	}

	// finding position of target function
	var start, end token.Pos
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Name.Name == d.env.FuncCode.FuncName {
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
					Specs: []ast.Spec{&ast.ValueSpec{Names: []*ast.Ident{ast.NewIdent("mu_apr")}, Type: ast.NewIdent("sync.Mutex")}},
				},
			}
			fn.Body.List = append([]ast.Stmt{stmt}, fn.Body.List...)
		}
		return true
	})

	// Create a buffer to hold the modified content
	var buf bytes.Buffer

	// Create a new printer.Config
	cfg := &printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}

	// Print the file to the buffer, modifying it as needed
	err = cfg.Fprint(&buf, fset, file)
	if err != nil {
		return err
	}

	// Convert the buffer to a string
	content := buf.String()
	var modifiedContent string

	re := regexp.MustCompile(fmt.Sprintf(`(var\s+%s|%s.*:=|return\s+.*%s)`, common, common, common))

	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, common) {
			found := re.FindAllString(line, -1)
			if len(found) == 0 {
				modifiedContent = fmt.Sprintf("%s\n%s\n%s\n%s", modifiedContent, "mu_apr.Lock()", line, "mu_apr.Unlock()")
			} else {
				modifiedContent = fmt.Sprintf("%s\n%s", modifiedContent, line)
			}
		} else {
			modifiedContent = fmt.Sprintf("%s\n%s", modifiedContent, line)
		}
	}

	// Write the modified content back to the file
	fd, err := os.Create(d.env.FuncCode.Path)
	if err != nil {
		return err
	}
	defer fd.Close()

	if _, err := fd.WriteString(modifiedContent); err != nil {
		return err
	}

	if err := utils.FixImports(d.env.FuncCode.Path); err != nil {
		return fmt.Errorf("can not fix imoprts of %s: %w", d.env.FuncCode.Path, err)
	}

	if err := utils.FormatGoFile(d.env.FuncCode.Path); err != nil {
		return fmt.Errorf("can not format go file: %w", err)
	}

	if err := d.env.FuncCode.UpdateCodeContentFromPath(); err != nil {
		return fmt.Errorf("can not update code content: %w", err)
	}

	return nil
}

func (d *DataRace) Revert() error {
	fd, err := os.Create(d.env.FuncCode.Path)
	if err != nil {
		return err
	}
	if _, err := fd.Write([]byte(d.oldFileContent)); err != nil {
		return err
	}

	if err := d.env.FuncCode.UpdateCodeContentFromPath(); err != nil {
		return fmt.Errorf("can not update code content: %w", err)
	}
	return nil
}

func parse(s string) []string {
	s = strings.ReplaceAll(s, "{", " ")
	s = strings.ReplaceAll(s, "}", " ")
	s = strings.ReplaceAll(s, "(", " ")
	s = strings.ReplaceAll(s, ")", " ")
	s = strings.ReplaceAll(s, ",", " ")
	s = strings.ReplaceAll(s, ".", " ")
	s = strings.ReplaceAll(s, "=", " ")
	s = strings.ReplaceAll(s, ":", " ")
	s = strings.ReplaceAll(s, "&", " ")
	s = strings.ReplaceAll(s, "%", " ")
	s = strings.ReplaceAll(s, "$", " ")
	s = strings.ReplaceAll(s, "#", " ")
	s = strings.ReplaceAll(s, "!", " ")
	s = strings.ReplaceAll(s, "*", " ")
	s = strings.ReplaceAll(s, "+", " ")
	s = strings.ReplaceAll(s, "\"", " ")
	s = strings.ReplaceAll(s, "'", " ")
	s = strings.ReplaceAll(s, "`", " ")
	s = strings.ReplaceAll(s, "\\", " ")
	s = strings.ReplaceAll(s, "|", " ")
	s = strings.ReplaceAll(s, "?", " ")
	s = strings.ReplaceAll(s, ">", " ")
	s = strings.ReplaceAll(s, "<", " ")
	s = strings.ReplaceAll(s, "~", " ")
	s = strings.ReplaceAll(s, "var", " ")
	s = strings.ReplaceAll(s, "int", " ")
	s = strings.ReplaceAll(s, "int64", " ")
	s = strings.ReplaceAll(s, "int32", " ")
	s = strings.ReplaceAll(s, "int16", " ")
	s = strings.ReplaceAll(s, "float32", " ")
	s = strings.ReplaceAll(s, "float16", " ")
	s = strings.ReplaceAll(s, "go", " ")
	s = strings.ReplaceAll(s, "return", " ")
	s = strings.ReplaceAll(s, "for", " ")
	s = strings.ReplaceAll(s, "func", " ")

	return strings.Fields(strings.TrimSpace(s))
}

func FindIntersection(lists [][]string) (string, error) {
	if len(lists) < 2 {
		return "", fmt.Errorf("there is only one line")
	}

	var newLists [][]string

	for _, list := range lists {
		if len(list) != 0 {
			newLists = append(newLists, list)
		}
	}

	var result string

	for _, str := range newLists[0] {
		ok := true
		for _, list := range newLists[0:] {
			if slices.Contains(list, str) {
				result = str
			} else {
				ok = false
			}
		}
		if ok {
			break
		}
	}

	return result, nil
}
