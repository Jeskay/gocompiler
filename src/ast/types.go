package ast

import (
	"gocompiler/src/parser"
	"gocompiler/src/tokens"
)

type Info struct {
	Defs map[parser.Ident]any

	Uses map[parser.Ident]any

	Implicits map[parser.Node]any
}

type Object interface {
	Parent() *Scope
	Pos() tokens.Position
	Name() string
	ToString()
	Type()
}
