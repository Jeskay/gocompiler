package ast

import (
	"fmt"
	"gocompiler/src/tokens"
)

type resolver struct {
	topScope         *Scope
	glScope          *Scope
	unresolved       []*Ident
	declarationError func(tokens.Position, string)
}

var unresolved = new(Object)

func ResolveFile(nodes []Node, declError func(tokens.Position, string)) *resolver {
	globalScope := NewScope(nil)
	r := &resolver{
		declarationError: declError,
		topScope:         globalScope,
		glScope:          globalScope,
	}
	for _, decl := range nodes {
		Walk(r, decl)
	}
	r.closeScope()
	i := 0
	for _, ident := range r.unresolved {
		ident.Obj = r.glScope.Lookup(ident.Name)
		if ident.Obj == nil {
			r.unresolved[i] = ident
			i++
		}
	}
	return r
}

func (r *resolver) resolve(ident *Ident, collectUnresolved bool) {
	if ident.Obj != nil {
		panic("identifier is already declared or resolved")
	}
	for s := r.topScope; s != nil; s = s.Outer {
		if obj := s.Lookup(ident.Name); obj != nil {
			if _, ok := obj.Decl.(*Ident); !ok {
				ident.Obj = obj
			}
			return
		}
	}
	if collectUnresolved {
		ident.Obj = unresolved
		r.unresolved = append(r.unresolved, ident)
	}
}

func (r *resolver) declare(decl, data any, scope *Scope, kind ObjectKind, idents ...*Ident) {
	for _, ident := range idents {
		obj := NewObject(ident.Name, kind)
		obj.Decl = decl
		obj.Data = data
		if _, ok := decl.(*Ident); !ok {
			ident.Obj = obj
		}
		if alt := scope.Insert(obj); alt != nil && r.declarationError != nil {
			prevDecl := ""
			if pos := alt.Pos(); pos.IsValid() {
				prevDecl = fmt.Sprintf("\n previous declaration at %v", pos)
			}
			r.declarationError(ident.Pos, fmt.Sprintf("%s redeclared in this block%s", ident.Name, prevDecl))
		}
	}
}

func (r *resolver) Visit(node Node) Visitor {
	switch n := node.(type) {
	case *Ident:
		r.resolve(n, true)
	case *FunctionLiteral:
		r.openScope(n.Type.Pos)
		defer r.closeScope()
		r.walkFunctionType(n.Type)
	case *SelectorExpression:
		Walk(r, n.X)
	case *StructType:
		r.openScope(n.Pos)
		defer r.closeScope()
		r.walkFieldList(n.Fields, Variable)
	case *FunctionType:
		r.openScope(n.Pos)
		defer r.closeScope()
		r.walkFunctionType(n)
	case *CompositeLiteral:
		if n.Type != nil {
			Walk(r, n.Type)
		}
		for _, e := range n.Elements {
			if keyValue, _ := e.(*KeyValueExpression); keyValue != nil {
				if ident, _ := keyValue.Key.(*Ident); ident != nil {
					r.resolve(ident, false)
				} else {
					Walk(r, keyValue.Key)
				}
				Walk(r, keyValue.Key)
			} else {
				Walk(r, e)
			}
		}
	case *AssignStatement:
		r.walkExpressions(n.Rhs)
		if n.Tok.Tok == tokens.DEFINE {
			r.shortVarDeclaration(n)
		} else {
			r.walkExpressions(n.Lhs)
		}
	case *BlockStatement:
		r.openScope(n.LbracePos)
		defer r.closeScope()
		r.walkStatements(n.List)
	case *IfStatement:
		r.openScope(n.Pos)
		defer r.closeScope()
		if n.Init != nil {
			Walk(r, n.Init)
		}
		Walk(r, n.Cond)
		Walk(r, n.Body)
		if n.Else != nil {
			Walk(r, n.Else)
		}
	case *ForStatement:
		r.openScope(n.Pos)
		defer r.closeScope()
		if n.Init != nil {
			Walk(r, n.Init)
		}
		if n.Cond != nil {
			Walk(r, n.Cond)
		}
		if n.Post != nil {
			Walk(r, n.Post)
		}
		Walk(r, n.Body)
	case *GenericDeclaration:
		switch n.Token {
		case tokens.CONST, tokens.VAR:
			for i, spec := range n.Specs {
				spec := spec.(*ValueSpec)
				var kind ObjectKind = Const
				if n.Token == tokens.VAR {
					kind = Variable
				}
				r.walkExpressions(spec.Values)
				if spec.Type != nil {
					Walk(r, spec.Type)
				}
				r.declare(spec, i, r.topScope, kind, spec.Names...)
			}
		case tokens.TYPE:
			for _, spec := range n.Specs {
				spec := spec.(*TypeSpec)
				r.declare(spec, nil, r.topScope, Type, spec.Name)
				if spec.TypeParams != nil {
					r.openScope(spec.Name.Pos)
					defer r.closeScope()
					r.walkTypeParams(spec.TypeParams)
				}
				Walk(r, spec.Type)
			}
		}
	case *FunctionDeclaration:
		r.openScope(n.Name.Pos)
		defer r.closeScope()
		if n.Type.TypeParams != nil {
			r.walkTypeParams(n.Type.TypeParams)
		}

		r.resolveList(n.Type.Params)
		r.resolveList(n.Type.Results)
		r.declareList(n.Type.Params, Variable)
		r.declareList(n.Type.Results, Variable)

		r.walkBody(n.Body)

		r.declare(n, nil, r.glScope, Function, n.Name)
	default:
		return r
	}
	return nil
}

func (r *resolver) openScope(pos tokens.Position) {
	r.topScope = NewScope(r.topScope)
}

func (r *resolver) closeScope() {
	r.topScope = r.topScope.Outer
}

func (r *resolver) walkFunctionType(typ *FunctionType) {
	r.resolveList(typ.Params)
	r.resolveList(typ.Results)
	r.declareList(typ.Params, Variable)
	r.declareList(typ.Results, Variable)
}

func (r *resolver) walkFieldList(list *FieldList, kind ObjectKind) {
	if list == nil {
		return
	}
	r.resolveList(list)
	r.declareList(list, kind)
}

func (r *resolver) walkExpressions(list []Expression) {
	for _, node := range list {
		Walk(r, node)
	}
}

func (r *resolver) walkStatements(list []Statement) {
	for _, stmt := range list {
		Walk(r, stmt)
	}
}
func (r *resolver) walkTypeParams(list *FieldList) {
	r.declareList(list, Type)
	r.resolveList(list)
}

func (r *resolver) walkBody(body *BlockStatement) {
	if body == nil {
		return
	}
	r.walkStatements(body.List)
}

func (r *resolver) resolveList(list *FieldList) {
	if list == nil {
		return
	}
	for _, f := range list.List {
		if f.Type != nil {
			Walk(r, f.Type)
		}
	}
}

func (r *resolver) declareList(list *FieldList, kind ObjectKind) {
	if list == nil {
		return
	}
	for _, f := range list.List {
		r.declare(f, nil, r.topScope, kind, f.Names...)
	}
}

func (r *resolver) shortVarDeclaration(decl *AssignStatement) {
	n := 0
	for _, x := range decl.Lhs {
		if ident, isIdent := x.(*Ident); isIdent {
			obj := NewObject(ident.Name, Variable)
			obj.Decl = decl
			ident.Obj = obj
			if alt := r.topScope.Insert(obj); alt != nil {
				ident.Obj = alt
			} else {
				n++
			}
		}
	}
	if n == 0 && r.declarationError != nil {
		r.declarationError(decl.Tok.Pos, "no new variables on the left side of :=")
	}
}
