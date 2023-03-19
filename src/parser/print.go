package parser

import (
	"strings"

	treePrinter "github.com/xlab/treeprint"
)

func PrintAST(nodes []Node) string {
	var result strings.Builder
	for _, node := range nodes {
		tree := treePrinter.New()
		node.printNode(tree)
		result.WriteString(tree.String())
	}
	return result.String()
}

func (i *Ident) printNode(tree treePrinter.Tree) {
	tree.AddNode(i.Name)
}

func (b *BasicLiteral) printNode(tree treePrinter.Tree) {
	tree.AddNode(b.Type.String() + " " + b.Value.LexString())
}

func (b *BinaryExpression) printNode(tree treePrinter.Tree) {
	t := tree.AddBranch(b.Operator.String())
	b.LeftX.printNode(t)
	b.RightX.printNode(t)
}

func (u *UnaryExpression) printNode(tree treePrinter.Tree) {
	t := tree.AddBranch(u.Operator.String())
	u.X.printNode(t)
}

func (d *FunctionDeclaration) printNode(tree treePrinter.Tree) {
	t := tree.AddBranch(d.Name.Name)
	body := t.AddBranch("body")
	typ := t.AddBranch("type")
	d.Body.printNode(body)
	d.Type.printNode(typ)
}

func (n *BlockStatement) printNode(tree treePrinter.Tree) {
	for _, stmt := range n.List {
		stmt.printNode(tree)
	}
}

func (n *FunctionType) printNode(tree treePrinter.Tree) {
	if n.TypeParams != nil {
		tParams := tree.AddBranch("type_params")
		n.TypeParams.printNode(tParams)
	}
	if n.Params != nil {
		params := tree.AddBranch("params")
		n.Params.printNode(params)
	}
	if n.Results != nil {
		results := tree.AddBranch("results")
		n.Results.printNode(results)
	}
}

func (n *FieldList) printNode(tree treePrinter.Tree) {
	for _, f := range n.List {
		f.printNode(tree)
	}
}

func (n *Field) printNode(tree treePrinter.Tree) {
	if n.Type == nil {
		return
	}
	t := tree.AddBranch("field")
	if len(n.Names) > 0 {
		names := t.AddBranch("names")
		for _, name := range n.Names {
			name.printNode(names)
		}
	}
	typ := t.AddBranch("type")

	n.Type.printNode(typ)
}

func (n *ArrayType) printNode(tree treePrinter.Tree) {
	t := tree.AddBranch("array")
	length := t.AddBranch("length")
	typ := t.AddBranch("type")
	n.Len.printNode(length)
	n.ElementType.printNode(typ)
}

func (n *StructType) printNode(tree treePrinter.Tree) {
	t := tree.AddBranch("struct")
	for _, field := range n.Fields {
		field.printNode(t)
	}
}

func (n *ValueSpec) printNode(tree treePrinter.Tree) {
	names := tree.AddBranch("names")
	typ := tree.AddBranch("type")
	values := tree.AddBranch("values")
	for _, name := range n.Names {
		name.printNode(names)
	}
	if n.Type != nil {
		n.Type.printNode(typ)
	}
	for _, value := range n.Values {
		value.printNode(values)
	}
}

func (n *TypeSpec) printNode(tree treePrinter.Tree) {
	spec := tree.AddBranch("spec")
	n.Name.printNode(spec.AddBranch("name"))
	n.Type.printNode(spec.AddBranch("type"))
	if n.TypeParams != nil {
		n.TypeParams.printNode(tree.AddBranch("type_params"))
	}
}

func (n *AssignStatement) printNode(tree treePrinter.Tree) {
	t := tree.AddBranch(n.Tok.LexString())
	l := t.AddBranch("left")
	for _, exp := range n.Lhs {
		exp.printNode(l)
	}
	r := t.AddBranch("right")
	for _, exp := range n.Rhs {
		exp.printNode(r)
	}
}

func (n *IfStatement) printNode(tree treePrinter.Tree) {
	t := tree.AddBranch("if")
	body := t.AddBranch("body")
	if n.Init != nil {
		init := t.AddBranch("init")
		n.Init.printNode(init)
	}
	if n.Else != nil {
		n.Else.printNode(t.AddBranch("else"))
	}
	n.Cond.printNode(t)
	n.Body.printNode(body)
}

func (n *ReturnStatement) printNode(tree treePrinter.Tree) {
	t := tree.AddBranch("return")
	for _, exp := range n.Results {
		exp.printNode(t)
	}
}

func (n *ForStatement) printNode(tree treePrinter.Tree) {
	t := tree.AddBranch("for")
	if n.Init != nil {
		n.Init.printNode(t.AddBranch("init"))
	}
	if n.Cond != nil {
		n.Cond.printNode(t.AddBranch("condition"))
	}
	if n.Post != nil {
		n.Post.printNode(t.AddBranch("post"))
	}
	body := t.AddBranch("body")
	n.Body.printNode(body)
}

func (n *ExpressionStatement) printNode(tree treePrinter.Tree) {
	n.X.printNode(tree)
}

func (n *IncDecStatement) printNode(tree treePrinter.Tree) {
	t := tree.AddBranch(n.Tok.ToString())
	n.X.printNode(t)
}

func (n *SelectorExpression) printNode(tree treePrinter.Tree) {
	t := tree.AddBranch("selector")
	name := t.AddBranch("name")
	n.Selector.printNode(name)
	method := t.AddBranch("method")
	n.X.printNode(method)
}

func (n *CallExpression) printNode(tree treePrinter.Tree) {
	fun := tree.AddBranch("method")
	n.Function.printNode(fun)
	args := fun.AddBranch("args")
	for _, arg := range n.Arguments {
		arg.printNode(args)
	}
}

func (n *DeclarationStatement) printNode(tree treePrinter.Tree) {
	t := tree.AddBranch("declaration")
	n.Decl.printNode(t)
}

func (n *GenericDeclaration) printNode(tree treePrinter.Tree) {
	t := tree.AddBranch(n.Token.String())
	for _, spec := range n.Specs {
		spec.printNode(t)
	}
}

func (n *IndexExpression) printNode(tree treePrinter.Tree) {
	t := tree.AddBranch("index_expression")
	n.X.printNode(t.AddBranch("name"))
	n.Index.printNode(t.AddBranch("index"))
}

func (n *IndexExpressions) printNode(tree treePrinter.Tree) {
	t := tree.AddBranch("index_expression")
	n.X.printNode(t.AddBranch("name"))
	indicies := t.AddBranch("indicies")
	for _, arg := range n.Indices {
		arg.printNode(indicies)
	}
}

func (n *CompositeLiteral) printNode(tree treePrinter.Tree) {
	t := tree.AddBranch("composite_literal")
	n.Type.printNode(t.AddBranch("type"))
	elems := t.AddBranch("elements")
	for _, elem := range n.Elements {
		elem.printNode(elems)
	}
}

func (n *KeyValueExpression) printNode(tree treePrinter.Tree) {
	t := tree.AddBranch("key_value")
	n.Key.printNode(t.AddBranch("key"))
	n.Value.printNode(t.AddBranch("value"))
}

func (n *FunctionLiteral) printNode(tree treePrinter.Tree) {
	t := tree.AddBranch("func")
	n.Type.printNode(t.AddBranch("type"))
	n.Body.printNode(t.AddBranch("body"))
}
