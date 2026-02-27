package snaps

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/kr/pretty"
)

type inlineSnapshotsLineMapping struct {
	// this map keeps a file inlineSnapshot call lines
	mapping map[string][]int
	sync.RWMutex
}

// AddLine
func (i *inlineSnapshotsLineMapping) AddLine(file string, line int) {
	i.Lock()
	defer i.Unlock()
	i.mapping[file] = append(i.mapping[file], line)
}

func (i *inlineSnapshotsLineMapping) AddFileIfNotExists(file string) bool {
	i.Lock()
	defer i.Unlock()
	if _, exists := i.mapping[file]; exists {
		return false
	}
	i.mapping[file] = make([]int, 0)
	return true
}

func (i *inlineSnapshotsLineMapping) GetLine(file string, index int) int {
	i.RLock()
	defer i.RUnlock()

	// this should not happen, returning -1 means we ignore it.
	if index >= len(i.mapping[file]) {
		return -1
	}

	line := i.mapping[file][index]

	return line
}

type inlineSnapshot *string

var (
	inlineSnapshotLineMapping = inlineSnapshotsLineMapping{
		mapping: make(map[string][]int),
		RWMutex: sync.RWMutex{},
	}
	errLocateCall = errors.New("cannot locate MatchInlineSnapshot call")
)

// Inline representation of snapshot
func Inline(s string) inlineSnapshot {
	return &s
}

/*
MatchInlineSnapshot compares a test value against an expected "inline snapshot" value.

Usage:

 1. On the first run, call with nil as the expected value:
    MatchInlineSnapshot(t, "mysnapshot", nil)
    This will record the current value as the snapshot.

 2. On subsequent runs, call with the snapshot value:
    MatchInlineSnapshot(t, "mysnapshot", snaps.Inline("mysnapshot"))
    This will verify that the value matches the stored snapshot.

An "inline snapshot" is a literal value stored directly in your test code, making it easy to update and review expected outputs.
*/
func (c *Config) MatchInlineSnapshot(t testingT, received any, inlineSnap inlineSnapshot) {
	t.Helper()

	matchInlineSnapshot(c, t, received, inlineSnap)
}

/*
MatchInlineSnapshot compares a test value against an expected "inline snapshot" value.

Usage:

 1. On the first run, call with nil as the expected value:
    MatchInlineSnapshot(t, "mysnapshot", nil)
    This will record the current value as the snapshot.

 2. On subsequent runs, call with the snapshot value:
    MatchInlineSnapshot(t, "mysnapshot", snaps.Inline("mysnapshot"))
    This will verify that the value matches the stored snapshot.

An "inline snapshot" is a literal value stored directly in your test code, making it easy to update and review expected outputs.
*/
func MatchInlineSnapshot(t testingT, received any, inlineSnap inlineSnapshot) {
	t.Helper()

	matchInlineSnapshot(&defaultConfig, t, received, inlineSnap)
}

func matchInlineSnapshot(c *Config, t testingT, received any, inlineSnap inlineSnapshot) {
	t.Helper()
	snapshot := c.takeInlineSnapshot(received)
	filename, line := baseCaller(1)

	// we should only register call positions if we are modifying the file and the file hasn't been registered yet.
	if (inlineSnap == nil || shouldUpdate(c.update)) &&
		inlineSnapshotLineMapping.AddFileIfNotExists(filename) {
		if err := registerInlineCallIdx(filename); err != nil {
			handleError(t, err)
			return
		}
	}

	if inlineSnap == nil {
		if !shouldCreate(c.update) {
			handleError(t, errSnapNotFound)
			return
		}

		if err := upsertInlineSnapshot(filename, line, snapshot); err != nil {
			handleError(t, err)
			return
		}

		t.Log(addedMsg)
		testEvents.register(added)
		return
	}

	diff := prettyDiff(*inlineSnap, snapshot, "", -1)
	if diff == "" {
		testEvents.register(passed)
		return
	}

	if !shouldUpdate(c.update) {
		handleError(t, diff)
		return
	}

	if err := upsertInlineSnapshot(filename, line, snapshot); err != nil {
		handleError(t, err)
		return
	}

	t.Log(updatedMsg)
	testEvents.register(updated)
}

func upsertInlineSnapshot(filename string, callerLine int, snapshot string) error {
	inlineSnapshotIdx := 0
	traverseError := errLocateCall

	fset, astFile, err := parseFileAst(filename)
	if err != nil {
		return err
	}

	traverseMatchInlineSnapshotAst(astFile, func(ce *ast.CallExpr) bool {
		if line := inlineSnapshotLineMapping.GetLine(filename, inlineSnapshotIdx); line == callerLine {
			ce.Args[2] = createInlineArgument(snapshot)

			// reset error as we found the caller
			traverseError = nil
			return false
		}

		inlineSnapshotIdx++
		// continue searching
		return true
	})
	if traverseError != nil {
		return traverseError
	}

	// Validate AST before writing
	var buf strings.Builder
	if err := format.Node(&buf, fset, astFile); err != nil {
		return fmt.Errorf(
			"invalid AST generated (snapshot contains problematic characters like backticks): %w",
			err,
		)
	}

	// Write to file
	file, err := os.OpenFile(filename, os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(buf.String()); err != nil {
		return err
	}

	return nil
}

func (c *Config) takeInlineSnapshot(received any) string {
	if c.serializer != nil {
		return c.serializer(received)
	}

	return pretty.Sprint(received)
}

// registerInlineCallIdx is expected to be called once per file and before getting modified
// it parses the file and registers all MatchInlineSnapshot call line numbers
func registerInlineCallIdx(filename string) error {
	fset, astFile, err := parseFileAst(filename)
	if err != nil {
		return err
	}

	traverseMatchInlineSnapshotAst(astFile, func(ce *ast.CallExpr) bool {
		inlineSnapshotLineMapping.AddLine(filename, fset.Position(ce.Pos()).Line)
		return true
	})

	return nil
}

/* AST Code */

func createInlineArgument(s string) ast.Expr {
	v := getInlineStringValue(s)

	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "snaps"},
			Sel: &ast.Ident{Name: "Inline"},
		},
		Args: []ast.Expr{&ast.BasicLit{
			Kind:  token.STRING,
			Value: v,
		}},
	}
}

func getInlineStringValue(s string) string {
	if strconv.Quote(s) != `"`+s+`"` {
		return fmt.Sprintf("`%s`", s)
	}

	return fmt.Sprintf("%q", s)
}

func traverseMatchInlineSnapshotAst(astFile *ast.File, fn func(*ast.CallExpr) bool) {
	breakEarly := false

	for _, decl := range astFile.Decls {
		if breakEarly {
			return
		}
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if !strings.HasPrefix(funcDecl.Name.Name, "Test") {
			continue
		}

		ast.Inspect(decl, func(n ast.Node) bool {
			if breakEarly {
				return false
			}
			callExpr, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			if selectorExpr.Sel.Name == "MatchInlineSnapshot" {
				if !fn(callExpr) {
					breakEarly = true
					return false
				}
			}

			return true
		})
	}
}

func parseFileAst(filename string) (*token.FileSet, *ast.File, error) {
	fileSet := token.NewFileSet()
	astFile, err := parser.ParseFile(
		fileSet,
		filename,
		nil,
		parser.ParseComments|parser.SkipObjectResolution,
	)
	if err != nil {
		return nil, nil, err
	}

	return fileSet, astFile, err
}
