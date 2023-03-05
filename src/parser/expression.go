package parser

import "gocompiler/src/tokens"

type Node interface{}

type Expression interface {
	Node
	exprNode()
}

type Statement interface {
	Node
	stmtNode()
}

type Spec interface {
	Node
	specNode()
}

type Declaration interface {
	Node
	declNode()
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

// spec nodes

type (
	ValueSpec struct {
		Names  []*Ident
		Type   Expression
		Values []Expression
	}

	TypeSpec struct {
		Name       *Ident
		TypeParams *FieldList
		AssignPos  tokens.Position
		Type       Expression
	}
)

// declaration nodes
type (
	FunctionDeclaration struct {
		Name *Ident
		Type *FunctionType
		Body *BlockStatement
	}
)

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
		Type *FunctionType
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
	FunctionType struct {
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
		Fields     []*Field
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

	ReturnStatement struct {
		Return  tokens.Position
		Results []Expression
	}

	IfStatement struct {
		Pos  tokens.Position
		Init Statement
		Cond Expression
		Body *BlockStatement
		Else Statement
	}

	ForStatement struct {
		Pos  tokens.Position
		Init Statement  // initialization statement; or nil
		Cond Expression // condition; or nil
		Post Statement  // post iteration statement; or nil
		Body *BlockStatement
	}

	AssignStatement struct {
		Lhs    []Expression
		TokPos tokens.Position // position of Tok
		Tok    tokens.Token    // assignment token, DEFINE
		Rhs    []Expression
	}

	IncDecStatement struct {
		X   Expression
		Pos tokens.Position // position of Tok
		Tok tokens.Token    // INC or DEC
	}

	ExpressionStatement struct {
		X Expression
	}
)

func (*Ident) exprNode()            {}
func (*BasicLiteral) exprNode()     {}
func (*UnaryExpression) exprNode()  {}
func (*BinaryExpression) exprNode() {}
func (*ArrayType) exprNode()        {}
func (*StructType) exprNode()       {}
func (*FunctionType) exprNode()     {}

func (*BlockStatement) stmtNode()      {}
func (*ReturnStatement) stmtNode()     {}
func (*IfStatement) stmtNode()         {}
func (*ForStatement) stmtNode()        {}
func (*AssignStatement) stmtNode()     {}
func (*IncDecStatement) stmtNode()     {}
func (*ExpressionStatement) stmtNode() {}

func (*ValueSpec) specNode() {}
func (*TypeSpec) specNode()  {}

func (*FunctionDeclaration) declNode() {}
