package lexer

import (
	"bufio"
	"io"
	"unicode"
)

type Token int

const (
	EOF = iota
	ILLEGAL
	IDENT
	INT
	SEMI

	ADD
	SUB
	MUL
	DIV
	ASSIGN
)

var tokens = []string{
	EOF:     "EOF",
	ILLEGAL: "ILLEGAL",
	IDENT:   "IDENT",
	INT:     "INT",
	SEMI:    "SEMI",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	DIV: "/",

	ASSIGN: "=",
}

func (t Token) String() string {
	return tokens[t]
}

type Position struct {
	Line   int
	Column int
}

type Lexer struct {
	position Position
	reader   *bufio.Reader
}

func NewLexer(reader io.Reader) *Lexer {
	return &Lexer{
		position: Position{Line: 1, Column: 0},
		reader:   bufio.NewReader(reader),
	}
}

func (l *Lexer) Lex() (Position, Token, string) {
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return l.position, EOF, ""
			}
			panic(err)
		}
		l.position.Column++
		switch r {
		case '\n':
			l.resetPosition()
		case ';':
			return l.position, SEMI, ";"
		case '+':
			return l.position, ADD, "+"
		case '-':
			return l.position, SUB, "-"
		case '*':
			return l.position, MUL, "*"
		case '/':
			return l.position, DIV, "/"
		case '=':
			return l.position, ASSIGN, "="
		default:
			if unicode.IsSpace(r) {
				continue
			} else if unicode.IsDigit(r) {
				startPos := l.position
				l.backup()
				lit := l.lexInt()
				return startPos, INT, lit
			} else if unicode.IsLetter(r) {
				startPos := l.position
				l.backup()
				lit := l.lexIdent()
				return startPos, IDENT, lit
			} else {
				return l.position, ILLEGAL, string(r)
			}
		}
	}
}

func (l *Lexer) resetPosition() {
	l.position.Line++
	l.position.Column = 0
}

func (l *Lexer) backup() {
	if err := l.reader.UnreadRune(); err != nil {
		panic(err)
	}
	l.position.Column--
}

func (l *Lexer) lexInt() string {
	var literal string
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return literal
			}
		}
		l.position.Column++
		if unicode.IsDigit(r) {
			literal += string(r)
		} else {
			l.backup()
			return literal
		}
	}
}

func (l *Lexer) lexIdent() string {
	var literal string
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return literal
			}
		}
		l.position.Column++
		if unicode.IsLetter(r) {
			literal += string(r)
		} else {
			l.backup()
			return literal
		}
	}
}
