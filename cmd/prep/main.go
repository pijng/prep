package main

import (
	"bytes"
	"fmt"
	"go/token"
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

	dstutil.Apply(f, func(c *dstutil.Cursor) bool {
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

		args := make([]string, len(argExpr.Args))
		identName := ""

		for idx, arg := range argExpr.Args {
			switch expr := arg.(type) {
			case *dst.Ident:
				args[idx] = expr.Name
				identName = expr.Name
			case *dst.BasicLit:
				value := expr.Value
				args[idx] = value
			}
		}

		argsString := strings.Join(args, ", ")
		if identName != "" {
			panic(fmt.Sprintf("cannot use identifier '%s' in function call '%s(%s)', use basic literals instead\n", identName, funcToCall, argsString))
		}

		fn, ok := funcs[funcToCall]
		if !ok {
			panic(fmt.Sprintf("cannot find func '%s' to eval", funcToCall))
		}

		_, err := cmpm.intr.Eval(fn)
		if err != nil {
			panic(fmt.Sprintf("cannot eval: %s", err))
		}

		call := fmt.Sprintf("%s%s.%s(%v)", PrepPrefix, funcToCall, funcToCall, argsString)
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
