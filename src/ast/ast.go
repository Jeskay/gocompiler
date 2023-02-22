package ast

import (
	"gocompiler/src/tokens"
)

type Node interface {
	Pos() tokens.Position
	End() tokens.Position
}

type Expr interface {
	Node
	expressionNode()
}

type BinaryExpression struct {
	left     Node
	right    Node
	operator tokens.TokenType
}
