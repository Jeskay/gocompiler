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
		node := p.parseExpression()
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

func (p *Parser) parseLiteral() (node BasicLiteral) {
	switch p.token.Tok {
	case tokens.INT, tokens.FLOAT, tokens.STRING, tokens.CHAR:
		node = BasicLiteral{Pos: p.token.Pos, Type: p.token.Tok, Value: p.token}
		p.next()
	default:
		panic("expected integer or string literal")
	}
	return
}

func (p *Parser) parseIdent() (node Ident) {
	if p.token.Tok == tokens.IDENT {
		node = Ident{Pos: p.token.Pos, Name: p.token.Lex.(string), Obj: p.token}
		p.next()
	} else {
		panic("expected identifier")
	}
	return
}

func (p *Parser) parsePrimaryExpression() (node Node) {
	return p.parseOperand()
}

func (p *Parser) parseOperand() (node Node) {
	switch p.token.Tok {
	case tokens.IDENT:
		return p.parseIdent()
	case tokens.LPAREN:
		p.next()
		node = p.parseExpression()
		p.expect(tokens.RPAREN)
		return
	default:
		return p.parseLiteral()
	}
}

func (p *Parser) parseUnaryExpression() (node Node) {
	switch p.token.Tok {
	case tokens.ADD, tokens.SUB, tokens.NOT:
		op := p.token
		p.next()
		return UnaryExpression{Pos: op.Pos, Operator: op.Tok, X: p.parseExpression()}
	default:
		return p.parsePrimaryExpression()
	}
}

func (p *Parser) parseOrExpression() (node Node) {
	node = p.parseAndExpression()
	for p.token.Tok == tokens.OR {
		operand := p.token
		p.next()
		right := p.parseAndExpression()
		node = BinaryExpression{Pos: operand.Pos, Operator: operand.Tok, LeftX: node, RightX: right}
	}
	return
}

func (p *Parser) parseAndExpression() (node Node) {
	node = p.parseComparisonExpression()
	for p.token.Tok == tokens.AND {
		operand := p.token
		p.next()
		right := p.parseComparisonExpression()
		node = BinaryExpression{Pos: operand.Pos, Operator: operand.Tok, LeftX: node, RightX: right}
	}
	return
}

func (p *Parser) parseComparisonExpression() (node Node) {
	node = p.parseAddExpression()
	for p.token.Tok == tokens.EQL {
		operand := p.token
		p.next()
		right := p.parseAddExpression()
		node = BinaryExpression{Pos: operand.Pos, Operator: operand.Tok, LeftX: node, RightX: right}
	}
	return
}

func (p *Parser) parseAddExpression() (node Node) {
	node = p.parseMulExpression()
	for p.token.Tok == tokens.ADD || p.token.Tok == tokens.SUB {
		operand := p.token
		p.next()
		right := p.parseMulExpression()
		node = BinaryExpression{Pos: operand.Pos, Operator: operand.Tok, LeftX: node, RightX: right}
	}
	return
}

func (p *Parser) parseMulExpression() (node Node) {
	node = p.parseUnaryExpression()
	for p.token.Tok == tokens.MUL || p.token.Tok == tokens.QUO {
		operand := p.token
		p.next()
		right := p.parseUnaryExpression()
		node = BinaryExpression{Pos: operand.Pos, Operator: operand.Tok, LeftX: node, RightX: right}
	}
	return
}

func (p *Parser) parseExpression() (node Node) {
	return p.parseOrExpression()
}

func (p *Parser) expect(tok tokens.TokenType) {
	if p.token.Tok != tok {
		panic("expected " + tok.String() + " but found " + p.token.Tok.String())
	}
	p.next()
}
