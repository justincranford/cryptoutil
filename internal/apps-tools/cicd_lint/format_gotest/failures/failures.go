package failures

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	cryptoutilCmdCicdCommon "cryptoutil/internal/apps-tools/cicd_lint/common"
	cryptoutilSharedMagic "cryptoutil/internal/shared/magic"
	cryptoutilSharedUtilFiles "cryptoutil/internal/shared/util/files"
)

type callEdit struct {
	start       int
	end         int
	replacement string
}

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
			logger.Log(fmt.Sprintf("Replaced %d testing.T failure calls in: %s", fixCount, filePath))
		}
	}

	return processed, modified, fixed, nil
}

func fixFile(filePath string) (bool, int, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, 0, fmt.Errorf("failed to read file: %w", err)
	}

	fileSet := token.NewFileSet()

	file, err := parser.ParseFile(fileSet, filePath, content, parser.ParseComments)
	if err != nil {
		return false, 0, fmt.Errorf("failed to parse file: %w", err)
	}

	hasErrorf, hasFatalf := failureCalls(file)
	if !hasErrorf && !hasFatalf {
		return false, 0, nil
	}

	if hasErrorf {
		content, err = ensureImport(content, "github.com/stretchr/testify/assert")
		if err != nil {
			return false, 0, err
		}
	}

	if hasFatalf {
		content, err = ensureImport(content, "github.com/stretchr/testify/require")
		if err != nil {
			return false, 0, err
		}
	}

	if hasErrorf || hasFatalf {
		content, err = ensureImportAs(content, "cryptoutil/internal/shared/magic", "cryptoutilSharedMagic")
		if err != nil {
			return false, 0, err
		}
	}

	fileSet = token.NewFileSet()

	file, err = parser.ParseFile(fileSet, filePath, content, parser.ParseComments)
	if err != nil {
		return false, 0, fmt.Errorf("failed to parse file after imports: %w", err)
	}

	aliases := packageAliases(file)

	edits := collectCallEdits(fileSet, file, content, aliases)
	if len(edits) == 0 {
		return false, 0, nil
	}

	sort.Slice(edits, func(left, right int) bool {
		return edits[left].start > edits[right].start
	})

	for _, edit := range edits {
		updated := make([]byte, 0, len(content)+len(edit.replacement))
		updated = append(updated, content[:edit.start]...)
		updated = append(updated, edit.replacement...)
		updated = append(updated, content[edit.end:]...)
		content = updated
	}

	content, err = format.Source(content)
	if err != nil {
		return false, len(edits), fmt.Errorf("failed to format file: %w", err)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return false, len(edits), fmt.Errorf("failed to stat file: %w", err)
	}

	if err := os.WriteFile(filePath, content, fileInfo.Mode().Perm()); err != nil {
		return false, len(edits), fmt.Errorf("failed to write file: %w", err)
	}

	return true, len(edits), nil
}

func failureCalls(file *ast.File) (bool, bool) {
	var hasErrorf, hasFatalf bool

	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		switch selector.Sel.Name {
		case cryptoutilSharedMagic.CICDTestMethodErrorf:
			if receiverIsT(selector.X) && len(call.Args) > 0 {
				hasErrorf = true
			}
		case cryptoutilSharedMagic.CICDTestMethodFatalf:
			if receiverIsT(selector.X) && len(call.Args) > 0 {
				hasFatalf = true
			}
		case cryptoutilSharedMagic.CICDTestMethodFailf, cryptoutilSharedMagic.CICDTestMethodFailNowf:
			if len(call.Args) < 2 || !receiverIsT(call.Args[0]) {
				return true
			}

			message, isMessage := call.Args[1].(*ast.BasicLit)
			if !isMessage || message.Kind != token.STRING {
				return true
			}

			messageValue, unquoteErr := strconv.Unquote(message.Value)
			if unquoteErr != nil {
				return true
			}

			switch messageValue {
			case cryptoutilSharedMagic.CICDTestErrorfFailureMessage:
				hasErrorf = true
			case cryptoutilSharedMagic.CICDTestFatalfFailureMessage:
				hasFatalf = true
			}
		}

		return true
	})

	return hasErrorf, hasFatalf
}

func receiverIsT(expression ast.Expr) bool {
	identifier, ok := expression.(*ast.Ident)

	return ok && identifier.Name == "t"
}

func packageAliases(file *ast.File) map[string]string {
	aliases := make(map[string]string)

	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}

		name := path.Base(importPath)
		if spec.Name != nil {
			name = spec.Name.Name
		}

		aliases[importPath] = name
	}

	return aliases
}

func collectCallEdits(fileSet *token.FileSet, file *ast.File, content []byte, aliases map[string]string) []callEdit {
	var edits []callEdit

	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		if selector.Sel.Name == cryptoutilSharedMagic.CICDTestMethodFailf || selector.Sel.Name == cryptoutilSharedMagic.CICDTestMethodFailNowf {
			if len(call.Args) < 2 || !receiverIsT(call.Args[0]) {
				return true
			}

			message, isMessage := call.Args[1].(*ast.BasicLit)
			if !isMessage || message.Kind != token.STRING {
				return true
			}

			messageValue, unquoteErr := strconv.Unquote(message.Value)
			if unquoteErr != nil {
				return true
			}

			constantName := ""

			switch messageValue {
			case cryptoutilSharedMagic.CICDTestErrorfFailureMessage:
				constantName = "CICDTestErrorfFailureMessage"
			case cryptoutilSharedMagic.CICDTestFatalfFailureMessage:
				constantName = "CICDTestFatalfFailureMessage"
			}

			if constantName == "" {
				return true
			}

			replacement := aliases["cryptoutil/internal/shared/magic"] + "." + constantName
			edits = append(edits, callEdit{
				start:       fileSet.Position(call.Args[1].Pos()).Offset,
				end:         fileSet.Position(call.Args[1].End()).Offset,
				replacement: replacement,
			})

			return true
		}

		if selector.Sel.Name != cryptoutilSharedMagic.CICDTestMethodErrorf && selector.Sel.Name != cryptoutilSharedMagic.CICDTestMethodFatalf || len(call.Args) == 0 {
			return true
		}

		if !receiverIsT(selector.X) {
			return true
		}

		importPath := "github.com/stretchr/testify/assert"
		methodName := cryptoutilSharedMagic.CICDTestMethodFailf
		failureMessageExpression := aliases["cryptoutil/internal/shared/magic"] + ".CICDTestErrorfFailureMessage"

		if selector.Sel.Name == cryptoutilSharedMagic.CICDTestMethodFatalf {
			importPath = "github.com/stretchr/testify/require"
			methodName = cryptoutilSharedMagic.CICDTestMethodFailNowf
			failureMessageExpression = aliases["cryptoutil/internal/shared/magic"] + ".CICDTestFatalfFailureMessage"
		}

		formatStart := fileSet.Position(call.Args[0].Pos()).Offset
		argumentsEnd := fileSet.Position(call.Args[len(call.Args)-1].End()).Offset
		receiverStart := fileSet.Position(selector.X.Pos()).Offset
		receiverEnd := fileSet.Position(selector.X.End()).Offset

		arguments := strings.TrimSpace(string(content[formatStart:argumentsEnd]))
		if call.Ellipsis.IsValid() {
			arguments += "..."
		}

		alias := aliases[importPath]
		replacement := fmt.Sprintf("%s.%s(%s, %s, %s)", alias, methodName, content[receiverStart:receiverEnd], failureMessageExpression, arguments)
		edits = append(edits, callEdit{
			start:       fileSet.Position(call.Pos()).Offset,
			end:         fileSet.Position(call.End()).Offset,
			replacement: replacement,
		})

		return true
	})

	return edits
}

func ensureImport(content []byte, importPath string) ([]byte, error) {
	return ensureImportAs(content, importPath, path.Base(importPath))
}

func ensureImportAs(content []byte, importPath, preferredAlias string) ([]byte, error) {
	fileSet := token.NewFileSet()

	file, err := parser.ParseFile(fileSet, "", content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file while adding import: %w", err)
	}

	for _, spec := range file.Imports {
		foundPath, unquoteErr := strconv.Unquote(spec.Path.Value)
		if unquoteErr == nil && foundPath == importPath {
			return content, nil
		}
	}

	alias := preferredAlias

	for _, spec := range file.Imports {
		foundPath, unquoteErr := strconv.Unquote(spec.Path.Value)
		if unquoteErr != nil {
			continue
		}

		importedName := path.Base(foundPath)
		if spec.Name != nil {
			importedName = spec.Name.Name
		}

		if importedName == alias {
			alias += "CICD"
		}
	}

	importSpec := strconv.Quote(importPath)
	if alias != path.Base(importPath) {
		importSpec = alias + " " + importSpec
	}

	for _, declaration := range file.Decls {
		importDeclaration, ok := declaration.(*ast.GenDecl)
		if !ok || importDeclaration.Tok != token.IMPORT {
			continue
		}

		if importDeclaration.Lparen.IsValid() {
			offset := fileSet.Position(importDeclaration.Rparen).Offset

			insertion := "\t" + importSpec + "\n"
			if offset > 0 && content[offset-1] != '\n' {
				insertion = ", " + importSpec
			}

			updated := make([]byte, 0, len(content)+len(insertion))
			updated = append(updated, content[:offset]...)
			updated = append(updated, insertion...)
			updated = append(updated, content[offset:]...)

			return updated, nil
		}

		start := fileSet.Position(importDeclaration.Pos()).Offset
		end := fileSet.Position(importDeclaration.End()).Offset
		oldSpec := strings.TrimSpace(string(content[start+len("import") : end]))
		replacement := "import (\n\t" + oldSpec + "\n\t" + importSpec + "\n)"
		updated := make([]byte, 0, len(content)+len(replacement))
		updated = append(updated, content[:start]...)
		updated = append(updated, replacement...)
		updated = append(updated, content[end:]...)

		return updated, nil
	}

	offset := fileSet.Position(file.Name.End()).Offset
	insertion := "\n\nimport " + importSpec
	updated := make([]byte, 0, len(content)+len(insertion))
	updated = append(updated, content[:offset]...)
	updated = append(updated, insertion...)
	updated = append(updated, content[offset:]...)

	return updated, nil
}
