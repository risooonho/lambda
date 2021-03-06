package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/faiface/lambda/ast"
	"github.com/faiface/lambda/machine"
	"github.com/faiface/lambda/parse"
)

func main() {
	eval := flag.String("eval", "", "evaluate a global")
	verbose := flag.Bool("v", false, "print all reduction steps")
	flag.Parse()

	globalNodes := map[string]ast.Node{
		"+":  &ast.Const{Value: &machine.IntBinOp{Type: machine.IntAdd}},
		"-":  &ast.Const{Value: &machine.IntBinOp{Type: machine.IntSub}},
		"*":  &ast.Const{Value: &machine.IntBinOp{Type: machine.IntMul}},
		"/":  &ast.Const{Value: &machine.IntBinOp{Type: machine.IntDiv}},
		"%":  &ast.Const{Value: &machine.IntBinOp{Type: machine.IntMod}},
		"==": &ast.Const{Value: &machine.IntCmpOp{Type: machine.IntEq}},
		"!=": &ast.Const{Value: &machine.IntCmpOp{Type: machine.IntNeq}},
		"<":  &ast.Const{Value: &machine.IntCmpOp{Type: machine.IntLess}},
		">":  &ast.Const{Value: &machine.IntCmpOp{Type: machine.IntMore}},
		"<=": &ast.Const{Value: &machine.IntCmpOp{Type: machine.IntLessEq}},
		">=": &ast.Const{Value: &machine.IntCmpOp{Type: machine.IntMoreEq}},
	}

	for _, path := range flag.Args() {
		file, err := os.Open(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		defer file.Close()
		nodes, err := parse.Definitions(path, file)
		if err != nil {
			err := err.(*parse.Error)
			fi := err.FileInfo
			msg := err.Msg
			exitWithError(fi, msg)
		}
		for name, node := range nodes {
			if globalNodes[name] != nil {
				fi := node.MetaInfo().(*parse.MetaInfo).FileInfo
				msg := fmt.Sprintf("'%s' already defined in another file", name)
				exitWithError(fi, msg)
			}
			globalNodes[name] = node
		}
	}

	globals, err := ast.CompileAll(globalNodes)
	if err != nil {
		err := err.(*ast.CompileError)
		fi := err.Node.MetaInfo().(*parse.MetaInfo).FileInfo
		msg := err.Msg
		exitWithError(fi, msg)
	}

	if *eval != "" {
		expr, ok := globals[*eval]
		if !ok {
			exitWithError(nil, fmt.Sprintf("eval: '%s' not defined", *eval))
		}
		machine.OneStepReduce = *verbose
		for !expr.IsNormal() {
			if *verbose {
				fmt.Println(show(expr))
				fmt.Println()
				fmt.Scanln()
			}
			expr = expr.Reduce()
		}
		fmt.Println(show(expr))
	}
}

func exitWithError(fi *parse.FileInfo, msg string) {
	if fi == nil {
		fmt.Fprintf(os.Stderr, "%s\n", msg)
	} else {
		fmt.Fprintf(os.Stderr, "%s:%d:%d: %s\n", fi.Filename, fi.Line, fi.Column, msg)
	}
	os.Exit(1)
}

func show(expr machine.Expr) string {
	repr := func(meta interface{}) string {
		mi, ok := meta.(*parse.MetaInfo)
		if !ok {
			return "(??)"
		}
		return mi.Name
	}
	return machine.ShowExpr(repr, expr)
}
