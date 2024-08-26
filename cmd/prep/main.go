package main

import (
	"bytes"
	"fmt"
	"go/token"
	"math"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/dstutil"
	"github.com/pijng/goinject"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

const PrepPrefix = "prep_"
const PrepPath = "github.com/pijng/prep"
const ComptimeName = "Comptime"

type ComptimeModifier struct {
	intr *interp.Interpreter
}

func main() {
	intr := interp.New(interp.Options{})
	err := intr.Use(stdlib.Symbols)
	if err != nil {
		panic(err)
	}

	cmpm := ComptimeModifier{intr: intr}

	goinject.Process(&cmpm)
}

func (cmpm *ComptimeModifier) Modify(f *dst.File, dec *decorator.Decorator, res *decorator.Restorer) *dst.File {
	funcs := collectFuncs(f, res)
	vars := collectVars(f)

	var parentFunc string
	dstutil.Apply(f, func(c *dstutil.Cursor) bool {
		funcDecl, ok := c.Node().(*dst.FuncDecl)
		if ok {
			parentFunc = funcDecl.Name.Name
		}

		callExpr, ok := c.Node().(*dst.CallExpr)
		if !ok {
			return true
		}

		funcIdent, isIdent := callExpr.Fun.(*dst.Ident)
		if !isIdent {
			return true
		}

		if funcIdent.Path != PrepPath && funcIdent.Name != ComptimeName {
			return true
		}

		argExpr, isExpr := callExpr.Args[0].(*dst.CallExpr)
		if !isExpr {
			return true
		}

		funcToCallIdent, isIdent := argExpr.Fun.(*dst.Ident)
		if !isIdent {
			return true
		}

		funcToCall := funcToCallIdent.Name

		origArgs := make([]string, len(argExpr.Args))
		args := make([]string, len(argExpr.Args))
		identName := ""

		for idx, arg := range argExpr.Args {
			switch expr := arg.(type) {
			case *dst.Ident:
				scopedName := fmt.Sprintf("%s_%s", parentFunc, expr.Name)
				v, ok := vars[scopedName]
				if !ok {
					origArgs[idx] = expr.Name
					identName = expr.Name

					continue
				}

				args[idx] = v
			case *dst.BasicLit:
				value := expr.Value
				args[idx] = value
				origArgs[idx] = value
			}
		}
		argsStr := strings.Join(args, ", ")
		origArgsStr := strings.Join(origArgs, ", ")

		if identName != "" {
			errStr := "unable to resolve comptime value '%s' in function call '%s(%s)'\n" +
				"argument to function being called at comptime must be comptime-known or represent a basic literal\n"
			panic(fmt.Sprintf(errStr, identName, funcToCall, origArgsStr))
		}

		fn, ok := funcs[funcToCall]
		if !ok {
			panic(fmt.Sprintf("cannot find func '%s' to eval", funcToCall))
		}

		_, err := cmpm.intr.Eval(fn)
		if err != nil {
			panic(fmt.Sprintf("cannot eval: %s", err))
		}

		call := fmt.Sprintf("%s%s.%s(%v)", PrepPrefix, funcToCall, funcToCall, argsStr)
		res, err := cmpm.intr.Eval(call)
		if err != nil {
			panic(fmt.Sprintf("cannot call: %s", err))
		}

		typeName := strings.ToUpper(res.Type().Name())
		tokenValue := token.Lookup(typeName)
		lit := &dst.BasicLit{Kind: tokenValue, Value: fmt.Sprint(res.Interface())}

		c.Replace(lit)

		return true
	}, nil)

	return f
}

func collectFuncs(f *dst.File, res *decorator.Restorer) map[string]string {
	funcs := make(map[string]string)

	for _, decl := range f.Decls {
		decl, isFunc := decl.(*dst.FuncDecl)
		if !isFunc {
			continue
		}

		var buf bytes.Buffer
		clonedDecl := dst.Clone(decl).(*dst.FuncDecl)
		funcF := &dst.File{
			Name:  dst.NewIdent(fmt.Sprintf("%s%s", PrepPrefix, clonedDecl.Name)),
			Decls: []dst.Decl{clonedDecl},
		}

		err := res.Fprint(&buf, funcF)
		if err != nil {
			panic(err)
		}

		funcs[decl.Name.Name] = buf.String()
	}

	return funcs
}

func collectVars(f *dst.File) map[string]string {
	vars := make(map[string]string)
	var parentFunc string

	dstutil.Apply(f, func(c *dstutil.Cursor) bool {
		funcDecl, ok := c.Node().(*dst.FuncDecl)
		if ok {
			parentFunc = funcDecl.Name.Name
		}

		assignStmt, ok := c.Node().(*dst.AssignStmt)
		if !ok {
			return true
		}

		for idx, lhs := range assignStmt.Lhs {
			ident, ok := lhs.(*dst.Ident)
			if !ok {
				continue
			}

			if len(assignStmt.Lhs) != len(assignStmt.Rhs) {
				op := float64(len(assignStmt.Rhs)) / float64(len(assignStmt.Lhs))
				idx = int(math.Floor(op))
			}

			rhs := assignStmt.Rhs[idx]
			lit, ok := rhs.(*dst.BasicLit)
			if !ok {
				continue
			}

			scopedName := fmt.Sprintf("%s_%s", parentFunc, ident.Name)
			vars[scopedName] = lit.Value
		}

		return true
	}, nil)

	return vars
}
