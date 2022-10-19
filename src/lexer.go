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
	FLOAT
	IMAG
	CHAR
	STRING
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

func (l *Lexer) Lex() (Position, Token, string, string) {
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return l.position, EOF, "", ""
			}
			panic(err)
		}
		l.position.Column++
		switch r {
		case '\n':
			l.nextLine()
		case ';':
			return l.position, SEMI, ";", string(r)
		case '+':
			return l.position, ADD, "+", string(r)
		case '-':
			return l.position, SUB, "-", string(r)
		case '*':
			return l.position, MUL, "*", string(r)
		case '/':
			return l.position, DIV, "/", string(r)
		case '=':
			return l.position, ASSIGN, "=", string(r)
		default:
			if unicode.IsSpace(r) {
				continue
			} else if unicode.IsDigit(r) {
				startPos := l.position
				l.backup()
				lex, lit := l.lexInt()
				return startPos, INT, lex, lit
			} else if unicode.IsLetter(r) {
				startPos := l.position
				l.backup()
				lex := l.lexIdent()
				return startPos, IDENT, lex, lex
			} else {
				return l.position, ILLEGAL, string(r), string(r)
			}
		}
	}
}

func (l *Lexer) nextLine() {
	l.position.Line++
	l.position.Column = 0
}

func (l *Lexer) backup() {
	if err := l.reader.UnreadRune(); err != nil {
		panic(err)
	}
	l.position.Column--
}

func (l *Lexer) lexInt() (string, string) {
	var lexem, base int64 = 0, 10
	var literal string = ""
	var i int64
	for i = 0;;i *= base {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return string(lexem), literal
			}
		}
		l.position.Column++
		if literal == "0" && r == 'x' {
			base = 16
			literal += string(r)
		} else if base == 16 && IsHex(r) {
			literal += string(r)
			lexem += RuneToInt(r) * i
		} else if IsDigit(r) {
			literal += string(r)
			lexem += int64(r - '0') * i
		} else {
			if base == 16 {
				l.backup()
			}
			l.backup()
			return string(lexem), literal
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
		if IsLetter(r) {
			literal += string(r)
		} else {
			l.backup()
			return literal
		}
	}
}
func RuneToInt(r rune) int64 {
	if IsDigit(r) {
		return int64(r - '0')
	} else if (r >= 'a' && r <= 'e') {
		return 10 + int64(r - 'a')
	} else if (r >= 'A' && r <= 'F') {
		return 10 + int64(r - 'A')
	} 
	panic("invalid hexadeciamal digit")
}
func IsDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func IsHex(r rune) bool {
	return IsDigit(r) || (r >= 'a' && r <= 'e') || (r >= 'A' && r <= 'E')
}

func IsLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}
