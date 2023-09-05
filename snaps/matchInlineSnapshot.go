package snaps

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"

	"github.com/kr/pretty"
)

type InlineSnapshot *string

// Inline representation of snapshot
func Inline(s string) InlineSnapshot {
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
func (c *config) MatchInlineSnapshot(t testingT, received interface{}, inlineSnap InlineSnapshot) {
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
func MatchInlineSnapshot(t testingT, received interface{}, inlineSnap InlineSnapshot) {
	t.Helper()

	matchInlineSnapshot(&defaultConfig, t, received, inlineSnap)
}

func matchInlineSnapshot(c *config, t testingT, received interface{}, inlineSnap InlineSnapshot) {
	t.Helper()
	snapshot := pretty.Sprint(received)
	filename, line := baseCaller(1)

	if inlineSnap == nil {
		if isCI {
			handleError(t, errSnapNotFound)
			return
		}

		if err := writeInlineSnapshot(filename, line, snapshot); err != nil {
			handleError(t, err)
		}
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

	if err := writeInlineSnapshot(filename, line, snapshot); err != nil {
		handleError(t, err)
		return
	}

	t.Log(updatedMsg)
	testEvents.register(updated)
}

func writeInlineSnapshot(filename string, line int, snapshot string) error {
	fset := token.NewFileSet()
	p, err := parser.ParseFile(
		fset,
		filename,
		nil,
		parser.ParseComments|parser.SkipObjectResolution,
	)
	if err != nil {
		return err
	}

	for _, decl := range p.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if !strings.HasPrefix(funcDecl.Name.Name, "Test") {
			continue
		}

		if !updateAST(decl, fset, line, snapshot) {
			continue
		}

		file, err := os.OpenFile(filename, os.O_TRUNC|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}
		defer file.Close()

		return printer.Fprint(file, fset, p)
	}

	return errors.New("cannot locate caller")
}

func updateAST(node ast.Node, fset *token.FileSet, line int, snapshot string) bool {
	var updated bool

	ast.Inspect(node, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		if selectorExpr.Sel.Name == "MatchInlineSnapshot" &&
			fset.Position(n.Pos()).Line == line {
			callExpr.Args[2] = createInlineArgument(snapshot)
			updated = true

			return false
		}

		return true
	})

	return updated
}

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
