package snaps

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path"
	"regexp"
	"strings"

	"github.com/gkampitakis/go-snaps/snaps/internal/colors"
)

var (
	skippedTests = newSyncSlice()
	skippedMsg   = colors.Sprint(colors.Yellow, skipSymbol+"Snapshot skipped\n")
)

// Wrapper of testing.Skip
//
// Keeps track which snapshots are getting skipped and not marked as obsolete.
func Skip(t testingT, args ...interface{}) {
	t.Helper()

	trackSkip(t)
	t.Skip(args...)
}

// Wrapper of testing.Skipf
//
// Keeps track which snapshots are getting skipped and not marked as obsolete.
func Skipf(t testingT, format string, args ...interface{}) {
	t.Helper()

	trackSkip(t)
	t.Skipf(format, args...)
}

// Wrapper of testing.SkipNow
//
// Keeps track which snapshots are getting skipped and not marked as obsolete.
func SkipNow(t testingT) {
	t.Helper()

	trackSkip(t)
	t.SkipNow()
}

func trackSkip(t testingT) {
	t.Helper()

	t.Log(skippedMsg)
	skippedTests.append(t.Name())
}

/*
This checks if the parent test is skipped,
or provided a 'runOnly' the testID is part of it

e.g

	func TestParallel (t *testing.T) {
		snaps.Skip(t)
		...
	}

Then every "child" test should be skipped
*/
func testSkipped(testID, runOnly string) bool {
	// testID form: Test.*/runName - 1
	testName := strings.Split(testID, " - ")[0]

	for _, name := range skippedTests.values {
		if testName == name || strings.HasPrefix(testName, name+"/") {
			return true
		}
	}

	matched, _ := regexp.MatchString(runOnly, testID)
	return !matched
}

func isFileSkipped(dir, filename, runOnly string) bool {
	// When a file is skipped through CLI with -run flag we can track it
	if runOnly == "" {
		return false
	}

	testFilePath := path.Join(dir, "..", strings.TrimSuffix(filename, snapsExt)+".go")
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
		matched, _ := regexp.MatchString(runOnly, funcDecl.Name.String())
		if matched {
			isSkipped = false
		}
	}

	return isSkipped
}
