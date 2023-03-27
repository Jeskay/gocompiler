package parser

import "fmt"

type Visitor interface {
	Visit(node Node) (w Visitor)
}

func walkIdentList(v Visitor, list []*Ident) {
	for _, x := range list {
		Walk(v, x)
	}
}

func walkExprList(v Visitor, list []Expression) {
	for _, x := range list {
		Walk(v, x)
	}
}

func walkStmtList(v Visitor, list []Statement) {
	for _, x := range list {
		Walk(v, x)
	}
}

func Walk(visitor Visitor, node Node) {
	var v Visitor
	if v = visitor.Visit(node); v == nil {
		return
	}

	switch n := node.(type) {
	case *Field:
		walkIdentList(v, n.Names)
		if n.Type != nil {
			Walk(v, n.Type)
		}
		if n.Tag != nil {
			Walk(v, n.Tag)
		}
	case *FieldList:
		for _, f := range n.List {
			Walk(v, f)
		}
	case *Ident, *BasicLiteral:
		//nothing
	case *FunctionLiteral:
		Walk(v, n.Type)
		Walk(v, n.Body)
	case *CompositeLiteral:
		if n.Type != nil {
			Walk(v, n.Type)
		}
		walkExprList(v, n.Elements)
	case *ParenExpressions:
		Walk(v, n.X)
	case *SelectorExpression:
		Walk(v, n.X)
		Walk(v, n.Selector)
	case *IndexExpression:
		Walk(v, n.X)
		Walk(v, n.Index)
	case *IndexExpressions:
		Walk(v, n.X)
		walkExprList(v, n.Indices)
	case *CallExpression:
		Walk(v, n.Function)
		walkExprList(v, n.Arguments)
	case *UnaryExpression:
		Walk(v, n.X)
	case *BinaryExpression:
		Walk(v, n.LeftX)
		Walk(v, n.RightX)
	case *KeyValueExpression:
		Walk(v, n.Key)
		Walk(v, n.Value)
	case *ArrayType:
		if n.Len != nil {
			Walk(v, n.Len)
		}
		Walk(v, n.ElementType)
	case *StructType:
		Walk(v, n.Fields)
	case *FunctionType:
		if n.TypeParams != nil {
			Walk(v, n.TypeParams)
		}
		if n.Params != nil {
			Walk(v, n.Params)
		}
		if n.Results != nil {
			Walk(v, n.Results)
		}
	case *DeclarationStatement:
		Walk(v, n.Decl)
	case *ExpressionStatement:
		Walk(v, n.X)
	case *IncDecStatement:
		Walk(v, n.X)
	case *AssignStatement:
		walkExprList(v, n.Lhs)
		walkExprList(v, n.Rhs)
	case *ReturnStatement:
		walkExprList(v, n.Results)
	case *BlockStatement:
		walkStmtList(v, n.List)
	case *IfStatement:
		if n.Init != nil {
			Walk(v, n.Init)
		}
		Walk(v, n.Cond)
		Walk(v, n.Body)
		if n.Else != nil {
			Walk(v, n.Else)
		}
	case *ForStatement:
		if n.Init != nil {
			Walk(v, n.Init)
		}
		if n.Cond != nil {
			Walk(v, n.Cond)
		}
		if n.Post != nil {
			Walk(v, n.Post)
		}
		Walk(v, n.Body)
	case *ValueSpec:
		walkIdentList(v, n.Names)
		if n.Type != nil {
			Walk(v, n.Type)
		}
		walkExprList(v, n.Values)
	case *TypeSpec:
		Walk(v, n.Name)
		if n.TypeParams != nil {
			Walk(v, n.TypeParams)
		}
		Walk(v, n.Type)
	case *GenericDeclaration:
		for _, s := range n.Specs {
			Walk(v, s)
		}
	case *FunctionDeclaration:
		Walk(v, n.Name)
		Walk(v, n.Type)
		if n.Body != nil {
			Walk(v, n.Body)
		}
	default:
		panic(fmt.Sprintf("Walk: unexpected node type %T", n))
	}
	v.Visit(nil)
}

type inspector func(Node) bool

func (f inspector) Visit(node Node) Visitor {
	if f(node) {
		return f
	}
	return nil
}

func Inspect(node Node, f func(Node) bool) {
	Walk(inspector(f), node)
}
