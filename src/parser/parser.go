package parser

import (
	"gocompiler/src/tokens"
)

type Parser struct {
	tokens  []tokens.Token
	current int
	token   tokens.Token
}

func NewParser(tokens []tokens.Token) *Parser {
	return &Parser{
		tokens:  tokens,
		current: 1,
		token:   tokens[0],
	}
}

func (p *Parser) Parse() (nodes []Node) {
	for p.current < len(p.tokens) {
		node := p.parseTopLevelDeclaration()
		nodes = append(nodes, node)
	}
	return
}

func (p *Parser) next() {
	if p.current < len(p.tokens) {
		p.token = p.tokens[p.current]
	}
	p.current++
}

func (p *Parser) parseLiteral() (node *BasicLiteral) {
	switch p.token.Tok {
	case tokens.INT, tokens.FLOAT, tokens.STRING, tokens.CHAR:
		node = &BasicLiteral{Pos: p.token.Pos, Type: p.token.Tok, Value: p.token}
		p.next()
	default:
		panic("expected integer or string literal")
	}
	return
}

func (p *Parser) parseIdent() (node *Ident) {
	if p.token.Tok == tokens.IDENT {
		node = &Ident{Pos: p.token.Pos, Name: p.token.Lex.(string), Obj: p.token}
		p.next()
	} else {
		panic("expected identifier")
	}
	return
}

func (p *Parser) parseIdentList() (node []*Ident) {
	node = append(node, p.parseIdent())
	for p.token.Tok == tokens.COMMA {
		p.next()
		node = append(node, p.parseIdent())
	}
	return
}

func (p *Parser) parseParamsList() (params []*Field) {
	params = append(params, p.parseParamDecl())
	for p.token.Tok == tokens.COMMA {
		p.next()
		params = append(params, p.parseParamDecl())
	}
	return
}

func (p *Parser) parseParameters(acceptTypeParams bool) (typeParams, params *FieldList) {
	if acceptTypeParams && p.token.Tok == tokens.LBRACK {
		opening := p.token.Pos
		p.next()
		list := p.parseParamsList()
		closing := p.expect(tokens.RBRACK).Pos
		typeParams = &FieldList{Opening: opening, List: list, Closing: closing}
	}
	opening := p.expect(tokens.LPAREN)
	var fields []*Field
	if p.token.Tok != tokens.RPAREN {
		fields = p.parseParamsList()
	}
	closing := p.expect(tokens.RPAREN)
	params = &FieldList{Opening: opening.Pos, List: fields, Closing: closing.Pos}
	return
}

func (p *Parser) parseResults() (node *FieldList) {
	if p.token.Tok == tokens.LPAREN {
		_, params := p.parseParameters(false)
		return params
	}
	typ := p.parseType()
	list := make([]*Field, 1)
	list[0] = &Field{Type: typ}
	return &FieldList{List: list}
}

func (p *Parser) parseFunctionType() *FunctionType {
	p.expect(tokens.FUNC)
	typeParams, params := p.parseParameters(true)
	results := p.parseResults()
	return &FunctionType{Pos: p.token.Pos, TypeParams: typeParams, Params: params, Results: results}
}

func (p *Parser) parseArrayType() (node *ArrayType) {
	length := p.parseExpression()
	p.expect(tokens.RBRACK)
	typ := p.parseType()
	return &ArrayType{Len: length, ElementType: typ, Post: p.token.Pos}
}

func (p *Parser) parseStructType() (node *StructType) {
	p.expect(tokens.STRUCT)
	p.expect(tokens.LBRACE)
	var list []*Field
	for p.token.Tok == tokens.IDENT || p.token.Tok == tokens.LPAREN {
		list = append(list, p.parseParamDecl())
	}
	p.expect(tokens.RBRACE)

	return &StructType{Pos: p.token.Pos, Fields: list}
}

func (p *Parser) parseParamDecl() *Field {
	params := p.parseIdentList()
	typ := p.parseType()
	return &Field{Names: params, Type: typ}
}

func (p *Parser) parseType() Expression {
	switch p.token.Tok {
	case tokens.IDENT:
		return p.parseIdent()
	case tokens.STRUCT:
		return p.parseStructType()
	case tokens.LBRACK:
		p.next()
		if p.token.Tok == tokens.RBRACK {
			p.next()
			typ := p.parseType()
			return &ArrayType{Len: nil, Post: p.token.Pos, ElementType: typ}
		}
		return p.parseArrayType()
	case tokens.FUNC:
		return p.parseFunctionType()
	default:
		return nil
	}
}

func (p *Parser) parseTypeArgs() []Expression {
	p.expect(tokens.LBRACK)
	var list []Expression
	list = append(list, p.parseType())
	for p.token.Tok == tokens.COMMA {
		list = append(list, p.parseType())
	}
	p.expect(tokens.RBRACK)
	return list
}
func (p *Parser) parseCall(function Expression) *CallExpression {
	lpos := p.expect(tokens.LPAREN).Pos
	var list []Expression
	for p.token.Tok != tokens.RPAREN && p.token.Tok != tokens.EOF {
		list = append(list, p.parseExpression())
		if p.token.Tok == tokens.RPAREN && p.token.Tok != tokens.EOF {
			break
		}
		p.next()
	}
	rpos := p.expect(tokens.RPAREN).Pos
	return &CallExpression{Function: function, LParenPos: lpos, RParenPos: rpos, Arguments: list}
}

func (p *Parser) parsePrimaryExpression(expr Expression) (node Expression) {
	if expr == nil {
		expr = p.parseOperand()
	}

	for n := 1; ; n++ {
		switch p.token.Tok {
		case tokens.PERIOD:
			p.next()
			switch p.token.Tok {
			case tokens.IDENT:
				name := p.parseIdent()
				expr = &SelectorExpression{X: expr, Selector: name}
			default:
				panic("exptected selector or type assertion")
			}
		case tokens.LPAREN:
			expr = p.parseCall(expr)
		case tokens.LBRACK:
			expr = p.parseIndexOrInstance(expr)
		default:
			return expr
		}
	}
}

func (p *Parser) parseIndexOrInstance(expr Expression) Expression {
	lpos := p.expect(tokens.LBRACK).Pos
	if p.token.Tok == tokens.RBRACK {
		panic("empty index, slice or index expressions are not permitted")
	}

	var args []Expression
	index := p.parseExpression()

	if p.token.Tok == tokens.COMMA {
		args = append(args, index)
		for p.token.Tok == tokens.COMMA {
			p.next()
			if p.token.Tok != tokens.RBRACK && p.token.Tok != tokens.EOF {
				args = append(args, p.parseType())
			}
		}
	}
	rpos := p.expect(tokens.RBRACK).Pos

	switch len(args) {
	case 0:
		return &IndexExpression{X: expr, LBracketPos: lpos, RBracketPos: rpos, Index: index}
	case 1:
		return &IndexExpression{X: expr, LBracketPos: lpos, RBracketPos: rpos, Index: args[0]}
	default:
		return &IndexExpressions{X: expr, Lbrack: lpos, Rbrack: rpos, Indices: args}
	}
}

func (p *Parser) parseGenericDeclaration(keyword tokens.TokenType) *GenericDeclaration {
	pos := p.expect(keyword).Pos
	var lpos, rpos tokens.Position
	var list []Spec
	if p.token.Tok == tokens.LPAREN {
		lpos = p.token.Pos
		p.next()
		for p.token.Tok != tokens.RPAREN && p.token.Tok != tokens.EOF {
			switch keyword {
			case tokens.VAR:
				list = append(list, p.parseVarSpec())
			case tokens.TYPE:
				list = append(list, p.parseTypeSpec())
			case tokens.CONST:
				list = append(list, p.parseConstSpec())
			}
		}
		rpos = p.expect(tokens.RPAREN).Pos
	} else {
		switch keyword {
		case tokens.VAR:
			list = append(list, p.parseVarSpec())
		case tokens.TYPE:
			list = append(list, p.parseTypeSpec())
		case tokens.CONST:
			list = append(list, p.parseConstSpec())
		}
	}

	return &GenericDeclaration{
		Token:     keyword,
		Pos:       pos,
		LParenPos: lpos,
		RParenPos: rpos,
		Specs:     list,
	}
}

func (p *Parser) parseOperand() (node Expression) {
	switch p.token.Tok {
	case tokens.IDENT:
		return p.parseIdent()
	case tokens.LPAREN:
		p.next()
		node = p.parseExpression()
		p.expect(tokens.RPAREN)
		return
	case tokens.INT, tokens.FLOAT, tokens.STRING, tokens.CHAR:
		return p.parseLiteral()
	case tokens.FUNC:
		return p.parseFunctionType()
	}
	return nil
}

func (p *Parser) parseStatement() Statement {
	switch p.token.Tok {
	case tokens.IF:
		return p.parseIfStatement()
	case tokens.FOR:
		return p.parseForStatement()
	case tokens.RETURN:
		return p.parseReturnStatement()
	case tokens.CONST, tokens.VAR, tokens.TYPE:
		return &DeclarationStatement{Decl: p.parseGenericDeclaration(p.token.Tok)}
	default:
		return p.parseSimpleStatement()
	}
}

func (p *Parser) parseSimpleStatement() Statement {
	expr := p.parseExpressionList()
	switch {
	case p.token.Tok >= tokens.ADD_ASSIGN && p.token.Tok < tokens.AND_NOT_ASSIGN:
		current := p.token
		p.next()
		y := p.parseExpressionList()
		return &AssignStatement{Lhs: expr, TokPos: current.Pos, Tok: current, Rhs: y}
	}
	switch p.token.Tok {
	case tokens.INC, tokens.DEC:
		statement := &IncDecStatement{X: expr[0], Pos: p.token.Pos, Tok: p.token}
		p.next()
		return statement
	}
	return &ExpressionStatement{X: expr[0]}
}

func (p *Parser) parseIfStatement() *IfStatement {
	pos := p.expect(tokens.IF).Pos
	exp := p.parseExpression()
	body := p.parseBlockStatement()
	var _else Statement
	if p.token.Tok == tokens.ELSE {
		p.next()
		switch p.token.Tok {
		case tokens.IF:
			_else = p.parseIfStatement()
		case tokens.LBRACE:
			_else = p.parseBlockStatement()
			p.expect(tokens.SEMICOLON)
		default:
			panic("expected if statement of block")
		}
	} else {
		p.expect(tokens.SEMICOLON)
	}
	return &IfStatement{Pos: pos, Cond: exp, Body: body, Else: _else}
}

func (p *Parser) parseForStatement() *ForStatement {
	pos := p.expect(tokens.FOR).Pos
	exp := p.parseExpression()
	body := p.parseBlockStatement()
	return &ForStatement{Pos: pos, Init: nil, Cond: exp, Post: nil, Body: body}
}

func (p *Parser) parseReturnStatement() *ReturnStatement {
	pos := p.expect(tokens.RETURN).Pos
	var expr []Expression
	if p.token.Tok != tokens.SEMICOLON && p.token.Tok != tokens.RBRACE {
		expr = p.parseExpressionList()
	}
	return &ReturnStatement{Return: pos, Results: expr}
}

func (p *Parser) parseStatementList() (list []Statement) {
	for p.token.Tok != tokens.RBRACE {
		list = append(list, p.parseStatement())
	}
	return
}

func (p *Parser) parseBlockStatement() *BlockStatement {
	begin := p.expect(tokens.LBRACE)
	list := p.parseStatementList()
	end := p.expect(tokens.RBRACE)
	return &BlockStatement{LbracePos: begin.Pos, List: list, RbracePos: end.Pos}
}

func (p *Parser) parseUnaryExpression() (node Expression) {
	switch p.token.Tok {
	case tokens.ADD, tokens.SUB, tokens.NOT:
		op := p.token
		p.next()
		return &UnaryExpression{Pos: op.Pos, Operator: op.Tok, X: p.parseExpression()}
	default:
		return p.parsePrimaryExpression(nil)
	}
}

func (p *Parser) parseBinaryExpression(expr Expression, prec tokens.TokenType) (node Expression) {

	if expr == nil {
		expr = p.parseUnaryExpression()
	}

	for {
		operand := p.token
		if operand.Tok < prec || operand.Tok > tokens.ELLIPSIS {
			break
		}
		p.next()

		right := p.parseBinaryExpression(nil, operand.Tok+1)
		expr = &BinaryExpression{Pos: operand.Pos, Operator: operand.Tok, LeftX: expr, RightX: right}
	}
	return expr
}

func (p *Parser) parseExpression() (node Expression) {
	return p.parseBinaryExpression(nil, tokens.ADD)
}

func (p *Parser) parseExpressionList() (list []Expression) {
	list = append(list, p.parseExpression())
	for p.token.Tok == tokens.COMMA {
		p.next()
		list = append(list, p.parseExpression())
	}
	return
}

func (p *Parser) parseConstSpec() *ValueSpec {
	idents := p.parseIdentList()
	var typ Expression
	var values []Expression
	if p.token.Tok != tokens.EOF && p.token.Tok != tokens.RPAREN {
		typ = p.parseType()
		if p.token.Tok == tokens.ASSIGN {
			p.next()
			values = p.parseExpressionList()
		}
	}
	p.expect(tokens.RPAREN)
	return &ValueSpec{Names: idents, Type: typ, Values: values}
}

func (p *Parser) parseVarSpec() *ValueSpec {
	idents := p.parseIdentList()
	var typ Expression
	var values []Expression
	if p.token.Tok != tokens.ASSIGN {
		typ = p.parseType()
	}
	if p.token.Tok == tokens.ASSIGN {
		p.next()
		values = p.parseExpressionList()
	}
	return &ValueSpec{Names: idents, Type: typ, Values: values}
}

func (p *Parser) parseFunctionDeclaration() *FunctionDeclaration {
	pos := p.expect(tokens.FUNC).Pos

	ident := p.parseIdent()
	typeParams, params := p.parseParameters(false)
	results := p.parseResults()
	var body *BlockStatement
	if p.token.Tok == tokens.LBRACE {
		body = p.parseBlockStatement()
	}

	return &FunctionDeclaration{
		Name: ident,
		Type: &FunctionType{
			Pos:        pos,
			TypeParams: typeParams,
			Params:     params,
			Results:    results,
		},
		Body: body,
	}
}

func (p *Parser) parseTypeSpec() (node Spec) {

	name := p.parseIdent()
	spec := &TypeSpec{Name: name}

	if p.token.Tok == tokens.LBRACK {
		p.next()
		spec.Type = p.parseArrayType()
	} else {
		if p.token.Tok == tokens.ASSIGN {
			spec.AssignPos = p.token.Pos
			p.next()
		}
		spec.Type = p.parseType()
	}
	return spec
}

func (p *Parser) parseTopLevelDeclaration() (node Node) {
	switch p.token.Tok {
	case tokens.CONST, tokens.VAR, tokens.TYPE:
		node = p.parseGenericDeclaration(p.token.Tok)
	case tokens.FUNC:
		node = p.parseFunctionDeclaration()
	}
	return
}

func (p *Parser) expect(tok tokens.TokenType) tokens.Token {
	if p.token.Tok != tok {
		panic("expected " + tok.String() + " but found " + p.token.Tok.String())
	}
	node := p.token
	p.next()
	return node
}

func (p *Parser) atComma(tok tokens.TokenType) {

}
