package snaps

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
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

func (i *inlineSnapshotsLineMapping) AddFile(file string) {
	i.Lock()
	defer i.Unlock()
	i.mapping[file] = make([]int, 0)
}

func (i *inlineSnapshotsLineMapping) isFileAdded(file string) bool {
	i.RLock()
	defer i.RUnlock()
	_, registered := i.mapping[file]

	return registered
}

func (i *inlineSnapshotsLineMapping) GetLine(file string, index int) int {
	i.RLock()
	defer i.RUnlock()

	return i.mapping[file][index]
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
MatchInlineSnapshot verifies the value matches the inline snapshot
First you call it with nil

	MatchInlineSnapshot(t, "mysnapshot", nil)

and it populates with the snapshot

	MatchInlineSnapshot(t, "mysnapshot", snaps.Inline("mysnapshot"))

the on every subsequent call it verifies the value matches the snapshot
*/
func (c *Config) MatchInlineSnapshot(t testingT, received interface{}, inlineSnap inlineSnapshot) {
	t.Helper()

	matchInlineSnapshot(c, t, received, inlineSnap)
}

/*
MatchInlineSnapshot verifies the value matches the inline snapshot
First you call it with nil

	MatchInlineSnapshot(t, "mysnapshot", nil)

and it populates with the snapshot

	MatchInlineSnapshot(t, "mysnapshot", snaps.Inline("mysnapshot"))

the on every subsequent call it verifies the value matches the snapshot
*/
func MatchInlineSnapshot(t testingT, received interface{}, inlineSnap inlineSnapshot) {
	t.Helper()

	matchInlineSnapshot(&defaultConfig, t, received, inlineSnap)
}

func matchInlineSnapshot(c *Config, t testingT, received interface{}, inlineSnap inlineSnapshot) {
	t.Helper()
	snapshot := pretty.Sprint(received)
	filename, line := baseCaller(1)

	// we should only register call positions if we are modifying the file and the file hasn't been registered yet.
	if (inlineSnap == nil || shouldUpdate(c.update)) &&
		!inlineSnapshotLineMapping.isFileAdded(filename) {
		if err := registerInlineCallIdx(filename); err != nil {
			handleError(t, err)
			return
		}
	}

	if inlineSnap == nil {
		if isCI {
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
	snapshotUpdated := false

	fset, astFile, err := parseFileAst(filename)
	if err != nil {
		return err
	}

	traverseMatchInlineSnapshotAst(astFile, func(ce *ast.CallExpr) bool {
		if inlineSnapshotLineMapping.GetLine(filename, inlineSnapshotIdx) == callerLine {
			ce.Args[2] = createInlineArgument(snapshot)
			snapshotUpdated = true
			return false
		}

		inlineSnapshotIdx++
		// continue searching
		return true
	})
	if !snapshotUpdated {
		return errLocateCall
	}

	file, err := os.OpenFile(filename, os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	return format.Node(file, fset, astFile)
}

// registerInlineCallIdx is expected to be called once per file and before getting modified
func registerInlineCallIdx(filename string) error {
	inlineSnapshotLineMapping.AddFile(filename)

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
	v := fmt.Sprintf("`%s`", s)
	if isSingleline(s) {
		v = fmt.Sprintf("%q", s)
	}

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
