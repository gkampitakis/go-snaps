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

func Inline(s string) InlineSnapshot {
	return &s
}

func MatchInlineSnapshot(t testingT, received interface{}, inlineSnap InlineSnapshot) {
	t.Helper()
	snapshot := pretty.Sprint(received)
	filename, line := baseCaller(1)

	if inlineSnap == nil {
		if isCI {
			handleError(t, errSnapNotFound)
			return
		}

		if err := injectSnapshot(filename, line, snapshot); err != nil {
			handleError(t, errSnapNotFound)
		}
		return
	}

	diff := prettyDiff(*inlineSnap, snapshot, "", -1)
	if diff == "" {
		testEvents.register(passed)
		return
	}

	if !shouldUpdateSingle(t.Name()) {
		handleError(t, diff)
		return
	}

	if err := injectSnapshot(filename, line, snapshot); err != nil {
		handleError(t, errSnapNotFound)
		return
	}

	t.Log(updatedMsg)
	testEvents.register(updated)
}

func injectSnapshot(filename string, line int, snapshot string) error {
	foundCaller := false
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
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if strings.HasPrefix(funcDecl.Name.Name, "Test") {
				ast.Inspect(decl, func(n ast.Node) bool {
					callExpr, ok := n.(*ast.CallExpr)
					if ok {
						selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
						if ok && selectorExpr.Sel.Name == "MatchInlineSnapshot" &&
							fset.Position(n.Pos()).Line == line {
							callExpr.Args[2] = createInlineArgument(snapshot)
							foundCaller = true
							return false
						}
					}

					return true
				})
			}

			if foundCaller {
				break
			}
		}
	}

	if !foundCaller {
		return errors.New("cannot locate caller")
	}

	file, err := os.OpenFile(filename, os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer file.Close()

	return printer.Fprint(file, fset, p)
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
