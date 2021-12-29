package snaps

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"strings"
	"testing"
)

var skippedTests = newSyncSlice()

// Wrapper of testing.Skip
func Skip(t *testing.T, args ...interface{}) {
	t.Helper()

	skippedTests.append(t.Name())
	t.Skip(args...)
}

// Wrapper of testing.Skipf
func Skipf(t *testing.T, format string, args ...interface{}) {
	t.Helper()

	skippedTests.append(t.Name())
	t.Skipf(format, args...)
}

// Wrapper of testing.SkipNow
func SkipNow(t *testing.T) {
	t.Helper()

	skippedTests.append(t.Name())
	t.SkipNow()
}

/*
	This checks if the parent test is skipped
	e.g
	func TestParallel (t *testing.T) {
		snaps.Skip()
		...
	}
	Then every "child" test should be skipped
*/
func testSkipped(testID, runOnly string) bool {
	if runOnly != "" && !strings.HasPrefix(testID, runOnly) {
		return true
	}

	for _, name := range skippedTests.values {
		if strings.HasPrefix(testID, name) {
			return true
		}
	}

	return false
}

func isFileSkipped(dir, filename, runOnly string) bool {
	// When a file is skipped through CLI with -run flag we can track it
	if runOnly == "" {
		return false
	}

	testFilePath := path.Join(dir, "..", strings.TrimSuffix(filename, ".snap")+".go")
	isSkipped := true

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testFilePath, nil, parser.ParseComments)
	if err != nil {
		return false
	}

	for _, decls := range file.Decls {
		funcDecl, ok := decls.(*ast.FuncDecl)

		if !ok {
			continue
		}

		// If the TestFunction is inside the file then it's not skipped
		if funcDecl.Name.String() == runOnly {
			isSkipped = false
		}
	}

	return isSkipped
}
