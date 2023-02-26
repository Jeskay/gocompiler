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

func (p *Parser) parsePrimaryExpression(expr Expression) (node Node) {
	if expr == nil {
		return p.parseOperand()
	}
	return expr
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
	case tokens.INT, tokens.FLOAT, tokens.STRING, tokens.CHAR:
		return p.parseLiteral()
	}
	return nil
}

func (p *Parser) parseUnaryExpression() (node Node) {
	switch p.token.Tok {
	case tokens.ADD, tokens.SUB, tokens.NOT:
		op := p.token
		p.next()
		return UnaryExpression{Pos: op.Pos, Operator: op.Tok, X: p.parseExpression()}
	default:
		return p.parsePrimaryExpression(nil)
	}
}

func (p *Parser) parseBinaryExpression(expr Expression, prec tokens.TokenType) (node Node) {

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
		expr = BinaryExpression{Pos: operand.Pos, Operator: operand.Tok, LeftX: expr, RightX: right}
	}
	return expr
}

func (p *Parser) parseExpression() (node Node) {
	return p.parseBinaryExpression(nil, tokens.ADD)
}

func (p *Parser) expect(tok tokens.TokenType) {
	if p.token.Tok != tok {
		panic("expected " + tok.String() + " but found " + p.token.Tok.String())
	}
	p.next()
}
