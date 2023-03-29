package ast

import "gocompiler/src/tokens"

type Scope struct {
	Outer   *Scope
	Objects map[string]*Object
}

func NewScope(outer *Scope) *Scope {
	return &Scope{outer, make(map[string]*Object, 4)}
}

func (s *Scope) Lookup(name string) *Object {
	return s.Objects[name]
}

func (s *Scope) Insert(obj *Object) (alt *Object) {
	if alt = s.Objects[obj.Name]; alt == nil {
		s.Objects[obj.Name] = obj
	}
	return
}

type Object struct {
	Kind ObjectKind
	Name string
	Decl any
	Data any
	Type any
}

func (o Object) Pos() tokens.Position {
	name := o.Name
	switch t := o.Decl.(type) {
	case *Field:
		for _, n := range t.Names {
			if n.Name == name {
				return n.Pos
			}
		}
	case *ValueSpec:
		for _, n := range t.Names {
			if n.Name == name {
				return n.Pos
			}
		}
	case *TypeSpec:
		if t.Name.Name == name {
			return t.Name.Pos
		}
	case *FunctionDeclaration:
		if t.Name.Name == name {
			return t.Name.Pos
		}
	case *AssignStatement:
		for _, x := range t.Lhs {
			if ident, isIdent := x.(*Ident); isIdent && ident.Name == name {
				return ident.Pos
			}
		}
	case *Scope:

	}
	return tokens.NoPosition
}

func NewObject(name string, kind ObjectKind) *Object {
	return &Object{Kind: kind, Name: name}
}

type ObjectKind int

const (
	Const = iota
	Type
	Variable
	Function
)

var objectKindStrings = [...]string{
	Const:    "const",
	Type:     "type",
	Variable: "var",
	Function: "func",
}

func (kind ObjectKind) String() string { return objectKindStrings[kind] }
