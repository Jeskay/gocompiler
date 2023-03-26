package ast

import "gocompiler/src/tokens"

type Scope struct {
	parent   *Scope
	children []*Scope
	elements map[string]Object
	pos, end tokens.Position
}

func NewScope(parent *Scope, pos, end tokens.Position) *Scope {
	s := &Scope{parent: parent, children: nil, elements: nil, pos: pos, end: end}
	if parent != nil {
		parent.children = append(parent.children, s)
	}
	return s
}

func (s *Scope) Parent() *Scope {
	return s.parent
}

func (s *Scope) Child(i int) *Scope {
	return s.children[i]
}

func (s *Scope) Names() []string {
	names := make([]string, len(s.elements))
	i := 0
	for name := range s.elements {
		names[i] = name
		i++
	}
	return names
}

func (s *Scope) Lookup(name string) Object {
	return s.elements[name]
}
