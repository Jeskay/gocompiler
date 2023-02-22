package parser

import "gocompiler/src/tokens"

type Node interface {
	Calc()
}

type Expression interface {
	Node
	Calc()
}

type Statement interface {
	Node
	stmtNode()
}

// Field is a field declaration in a struct type
type Field struct {
	Names []*Ident
	Type  Expression
	Tag   *BasicLiteral
}

// FieldList is a list of fields
type FieldList struct {
	Opening tokens.Position
	List    []*Field
	Closing tokens.Position
}

func (f *FieldList) NumFields() int {
	n := 0
	if f != nil {
		for _, g := range f.List {
			m := len(g.Names)
			if m == 0 {
				m = 1
			}
			n += m
		}
	}
	return n
}

// expression nodes
type (
	Ident struct {
		Pos  tokens.Position
		Name string
		Obj  any
	}

	Ellipsis struct {
		Pos tokens.Position
		Elt Expression
	}

	BasicLiteral struct {
		Pos   tokens.Position
		Type  tokens.TokenType
		Value tokens.Token
	}

	FuncLiteral struct {
		Type *FuncType
		Body *BlockStatement
	}

	CompositeLitral struct {
		Type       Expression
		LbracePos  tokens.Position
		RbracePos  tokens.Position
		Elements   []Expression
		Incomplete bool
	}

	ParenExpressions struct {
		X        Expression
		Selector *Ident
	}

	IndexExpressions struct {
		X       Expression
		Lbrace  tokens.Position
		Rbrace  tokens.Position
		Indices []Expression
	}

	UnaryExpression struct {
		Pos      tokens.Position
		Operator tokens.TokenType
		X        Expression
	}

	BinaryExpression struct {
		Pos      tokens.Position
		Operator tokens.TokenType
		LeftX    Expression
		RightX   Expression
	}
)

// type-specific expression nodes
type (
	FuncType struct {
		Pos        tokens.Position
		TypeParams *FieldList
		Params     *FieldList
		Results    *FieldList
	}

	ArrayType struct {
		Post        tokens.Position
		Len         Expression
		ElementType Expression
	}

	StructType struct {
		Pos        tokens.Position
		Fields     *FieldList
		Incomplete bool
	}
)

// statements
type (
	BlockStatement struct {
		LbracePos tokens.Position
		List      []Statement
		RbracePos tokens.Position
	}
)

func Calc(n Node) {

}

func (n BasicLiteral) Calc() {

}

func (n UnaryExpression) Calc() {

}

func (n BinaryExpression) Calc() {

}

func (n Ident) Calc() {

}
