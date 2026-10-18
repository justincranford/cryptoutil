package parallel

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilFiles "cryptoutil/internal/shared/util/files"
)

func Fix(logger *cryptoutilCmdCicdCommon.Logger, rootDir string) (int, int, int, error) {
	filesByExtension, err := cryptoutilSharedUtilFiles.ListAllFiles(rootDir)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to list files: %w", err)
	}

	testFiles := cryptoutilCmdCicdCommon.FilterFilesForCommand(filesByExtension["go"], "format-go-test")

	var processed, modified, fixed int

	for _, filePath := range testFiles {
		if !strings.HasSuffix(filePath, "_test.go") {
			continue
		}

		changed, fixCount, fixErr := fixFile(filePath)
		if fixErr != nil {
			return processed, modified, fixed, fmt.Errorf("failed to process %s: %w", filePath, fixErr)
		}

		processed++

		if changed {
			modified++
			fixed += fixCount
			logger.Log(fmt.Sprintf("Added t.Parallel() to %d tests in: %s", fixCount, filePath))
		}
	}

	return processed, modified, fixed, nil
}

func fixFile(filePath string) (bool, int, error) {
	return fixFileWithOperations(filePath, parallelFileOperations{
		readFile:  os.ReadFile,
		stat:      os.Stat,
		writeFile: os.WriteFile,
	})
}

type parallelFileOperations struct {
	readFile  func(string) ([]byte, error)
	stat      func(string) (os.FileInfo, error)
	writeFile func(string, []byte, os.FileMode) error
}

func fixFileWithOperations(filePath string, fileOperations parallelFileOperations) (bool, int, error) {
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return false, 0, fmt.Errorf("failed to parse file: %w", err)
	}

	type insertion struct {
		offset    int
		parameter string
	}

	var insertions []insertion

	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv != nil || !strings.HasPrefix(function.Name.Name, "Test") || function.Body == nil {
			continue
		}

		parameterName := testingTParameter(function)
		if parameterName == "" || hasParallelCall(function.Body, parameterName) || hasSequentialComment(file, fset, function) {
			continue
		}

		insertions = append(insertions, insertion{
			offset:    fset.Position(function.Body.Lbrace).Offset + 1,
			parameter: parameterName,
		})
	}

	if len(insertions) == 0 {
		return false, 0, nil
	}

	sort.Slice(insertions, func(left, right int) bool {
		return insertions[left].offset < insertions[right].offset
	})

	content, err := fileOperations.readFile(filePath)
	if err != nil {
		return false, len(insertions), fmt.Errorf("failed to read file: %w", err)
	}

	formattedContent := make([]byte, 0, len(content)+len(insertions)*len("\n\tt.Parallel()\n"))

	previousOffset := 0
	for _, item := range insertions {
		formattedContent = append(formattedContent, content[previousOffset:item.offset]...)
		formattedContent = append(formattedContent, []byte("\n\t"+item.parameter+".Parallel()\n")...)
		previousOffset = item.offset
	}

	formattedContent = append(formattedContent, content[previousOffset:]...)

	formattedContent, err = format.Source(formattedContent)
	if err != nil {
		return false, len(insertions), fmt.Errorf("failed to format file: %w", err)
	}

	fileInfo, err := fileOperations.stat(filePath)
	if err != nil {
		return false, len(insertions), fmt.Errorf("failed to stat file: %w", err)
	}

	if err := fileOperations.writeFile(filePath, formattedContent, fileInfo.Mode().Perm()); err != nil {
		return false, len(insertions), fmt.Errorf("failed to write file: %w", err)
	}

	return true, len(insertions), nil
}

func testingTParameter(function *ast.FuncDecl) string {
	if function.Type.Params == nil {
		return ""
	}

	for _, field := range function.Type.Params.List {
		pointer, ok := field.Type.(*ast.StarExpr)
		if !ok {
			continue
		}

		selector, ok := pointer.X.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "T" {
			continue
		}

		packageName, ok := selector.X.(*ast.Ident)
		if !ok || packageName.Name != "testing" || len(field.Names) == 0 {
			continue
		}

		return field.Names[0].Name
	}

	return ""
}

func hasParallelCall(body *ast.BlockStmt, parameterName string) bool {
	found := false

	ast.Inspect(body, func(node ast.Node) bool {
		if _, isNestedFunction := node.(*ast.FuncLit); isNestedFunction {
			return false
		}

		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Parallel" {
			return true
		}

		parameter, ok := selector.X.(*ast.Ident)
		if ok && parameter.Name == parameterName {
			found = true
		}

		return true
	})

	return found
}

func hasSequentialComment(file *ast.File, fset *token.FileSet, function *ast.FuncDecl) bool {
	functionLine := fset.Position(function.Pos()).Line
	for _, group := range file.Comments {
		commentEndLine := fset.Position(group.End()).Line
		if functionLine > commentEndLine && functionLine-commentEndLine <= cryptoutilSharedMagic.SequentialCommentWindow && strings.Contains(group.Text(), "Sequential:") {
			return true
		}
	}

	return false
}
