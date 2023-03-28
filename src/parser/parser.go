package parser

import (
	"gocompiler/src/lexer"
	"gocompiler/src/tokens"
	"io"
)

type Parser struct {
	token         tokens.Token
	lexerInstance *lexer.Lexer
}

func NewParser(reader io.Reader) *Parser {
	lexer := lexer.NewLexer(reader)
	parser := &Parser{
		lexerInstance: lexer,
		token:         lexer.GetLexem(),
	}
	return parser
}

func (p *Parser) Parse() (nodes []Node) {
	for p.token.Tok != tokens.EOF {
		node := p.parseTopLevelDeclaration()
		nodes = append(nodes, node)
	}
	return
}

func (p *Parser) next() {
	p.token = p.lexerInstance.GetLexem()
	if p.token.Tok == tokens.COMMENT {
		p.next()
	}
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
		node = &Ident{Pos: p.token.Pos, Name: p.token.Lex.(string)}
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
	opening := p.expect(tokens.LBRACE).Pos
	var list []*Field
	for p.token.Tok == tokens.IDENT || p.token.Tok == tokens.LPAREN {
		list = append(list, p.parseParamDecl())
	}
	closing := p.expect(tokens.RBRACE).Pos

	return &StructType{Pos: p.token.Pos, Fields: &FieldList{List: list, Opening: opening, Closing: closing}}
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

func (p *Parser) parseValue() Expression {
	if p.token.Tok == tokens.LBRACE {
		return p.parseLiteralValue(nil)
	}
	return p.parseExpression()
}

func (p *Parser) parseElement() Expression {
	key := p.parseValue()
	if p.token.Tok == tokens.COLON {
		colon := p.token.Pos
		p.next()
		key = &KeyValueExpression{Key: key, ColonPos: colon, Value: p.parseValue()}
	}
	return key
}

func (p *Parser) parseElementList() (list []Expression) {
	for p.token.Tok != tokens.RBRACE && p.token.Tok != tokens.EOF {
		list = append(list, p.parseElement())
		if p.token.Tok == tokens.RBRACE && p.token.Tok != tokens.EOF {
			break
		}
		p.next()
	}
	return
}

func (p *Parser) parseLiteralValue(typ Expression) Expression {
	lpos := p.expect(tokens.LBRACE).Pos
	var elements []Expression
	if p.token.Tok != tokens.RBRACE {
		elements = p.parseElementList()
	}
	rpos := p.expect(tokens.RBRACE).Pos
	return &CompositeLiteral{Type: typ, LbracePos: lpos, RbracePos: rpos, Elements: elements}
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
		case tokens.LBRACE:
			switch expr.(type) {
			case *ArrayType, *StructType:
				expr = p.parseLiteralValue(expr)
			case *Ident, *SelectorExpression:
				expr = p.parseLiteralValue(expr)
			default:
				return expr
			}
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
		typ := p.parseFunctionType()
		if p.token.Tok != tokens.LBRACE {
			return typ
		}
		body := p.parseBlockStatement()
		return &FunctionLiteral{Type: typ, Body: body}
	case tokens.STRUCT:
		return p.parseStructType()
	case tokens.LBRACK:
		p.expect(tokens.LBRACK)
		return p.parseArrayType()
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
	case tokens.LBRACE:
		return p.parseBlockStatement()
	default:
		return p.parseSimpleStatement()
	}
}

func (p *Parser) parseSimpleStatement() Statement {
	expr := p.parseExpressionList()
	switch {
	case p.token.Tok == tokens.DEFINE, p.token.Tok >= tokens.ADD_ASSIGN && p.token.Tok < tokens.AND_NOT_ASSIGN:
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
	if p.token.Tok == tokens.LBRACE {
		panic("missing condition in if statement")
	}
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
		default:
			panic("expected if statement of block")
		}
	}
	return &IfStatement{Pos: pos, Cond: exp, Body: body, Else: _else}
}

func (p *Parser) parseForStatement() *ForStatement {
	pos := p.expect(tokens.FOR).Pos
	var stmt1, stmt2, stmt3 Statement
	if p.token.Tok != tokens.LBRACE {
		if p.token.Tok != tokens.SEMICOLON {
			stmt2 = p.parseSimpleStatement()
		}
		if p.token.Tok == tokens.SEMICOLON {
			p.next()
			stmt1 = stmt2
			stmt2 = nil
			if p.token.Tok != tokens.SEMICOLON {
				stmt2 = p.parseSimpleStatement()
			}
			p.optionalSemi()
			if p.token.Tok != tokens.LBRACE {
				stmt3 = p.parseSimpleStatement()
			}
		}
	}
	body := p.parseBlockStatement()
	return &ForStatement{Pos: pos, Init: stmt1, Cond: p.toExpr(stmt2, "boolean expression"), Post: stmt3, Body: body}
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
		if operand.Tok < prec || operand.Tok >= tokens.DEFINE || operand.Tok == tokens.INC || operand.Tok == tokens.DEC {
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
		if p.token.Tok != tokens.ASSIGN {
			typ = p.parseType()
		}
		if p.token.Tok == tokens.ASSIGN {
			p.next()
			values = p.parseExpressionList()
		}
	}
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
		panic(p.token.Pos.ToString() + " expected " + tok.String() + " but found " + p.token.Tok.String())
	}
	node := p.token
	p.next()
	return node
}

func (p *Parser) optionalSemi() {
	if p.token.Tok != tokens.RPAREN && p.token.Tok != tokens.RBRACE {
		if p.token.Tok == tokens.SEMICOLON {
			p.next()
		} else {
			panic(p.token.Pos.ToString() + " expected ;")
		}
	}
}

func (p *Parser) toExpr(s Statement, expected string) Expression {
	if s == nil {
		return nil
	}
	if expr, isExpr := s.(*ExpressionStatement); isExpr {
		return expr.X
	}
	if _, isAssign := s.(*AssignStatement); isAssign {
		panic(p.token.Pos.ToString() + " expected " + expected + "but found assignment")
	} else {
		panic(p.token.Pos.ToString() + " expected " + expected + "but found simple statement")
	}
	//return &BadExpression{From: p.token.Pos, To: p.token.Pos}
}
