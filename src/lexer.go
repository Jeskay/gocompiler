package lexer

import (
	"bufio"
	"io"
	"unicode"
)

type Position struct {
	Line   int
	Column int
}

type Lexer struct {
	position Position
	reader   *bufio.Reader
	buffer   *Buffer
}

func NewLexer(reader io.Reader) *Lexer {
	return &Lexer{
		position: Position{Line: 1, Column: 0},
		reader:   bufio.NewReader(reader),
		buffer:   NewBuffer(5),
	}
}

func (l *Lexer) Lex() (Position, Token, string, string) {
	for {
		r, err := l.readNext()
		if err != nil {
			if err == io.EOF {
				return l.position, EOF, "", ""
			}
			panic(err)
		}
		switch r {
		case '\n':
			l.nextLine()
		case ';':
			return l.position, SEMICOLON, ";", string(r)
		case '+':
			return l.position, ADD, "+", string(r)
		case '-':
			return l.position, SUB, "-", string(r)
		case '*':
			return l.position, MUL, "*", string(r)
		case '/':
			startPos := l.position
			l.backup()
			token, lex, lit := l.lexComment()
			return startPos, token, lex, lit
		case '=':
			return l.position, ASSIGN, "=", string(r)
		case '\'':
			startPos := l.position
			l.backup()
			token, lex, lit := l.lexChar()
			return startPos, token, lex, lit
		default:
			if unicode.IsSpace(r) {
				continue
			} else if IsDigit(r) {
				startPos := l.position
				l.backup()
				lex, lit := l.lexInt()
				return startPos, INT, lex, lit
			} else if IsLetter(r) {
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
	if !l.buffer.isEmpty {
		l.buffer.position--
	}
	l.position.Column--
}

func (l *Lexer) readNext() (symbol rune, err error) {
	if !l.buffer.CurrentAtHead() {
		l.buffer.position++
		l.position.Column++
		return l.buffer.GetCurrent(), nil
	}
	r, _, err := l.reader.ReadRune()
	if err != nil {
		return 0, err
	}
	l.position.Column++
	l.buffer.Push(r)
	if !l.buffer.CurrentAtHead() {
		l.buffer.position++
	}
	symbol = l.buffer.GetCurrent()
	return
}

func (l *Lexer) lexInt() (string, string) {
	var lexem, base int64 = 0, 10
	var literal string = ""
	for {
		r, err := l.readNext()
		if err != nil {
			if err == io.EOF {
				return intToString(lexem), literal
			}
		}

		if literal == "0" {
			switch r {
			case 'x', 'X':
				base = 16
				literal += string(r)
				continue
			case 'o', 'O':
				base = 8
				literal += string(r)
				continue
			case 'b', 'B':
				base = 2
				literal += string(r)
				continue
			default:
				base = 8
			}
		}
		if RuneInBase(base, r) {
			literal += string(r)
			lexem = RuneToInt(r) + lexem*base
		} else {
			if literal == "0x" {
				literal = "0"
				l.backup()
			}
			l.backup()
			return intToString(lexem), literal
		}
	}
}

func (l *Lexer) lexIdent() string {
	var literal string
	for {
		r, err := l.readNext()
		if err != nil {
			if err == io.EOF {
				return literal
			}
		}
		if IsLetter(r) || (len(literal) > 0 && IsDigit(r)) {
			literal += string(r)
		} else {
			l.backup()
			return literal
		}
	}
}

func (l *Lexer) lexComment() (token Token, lexem string, literal string) {
	comment := ""
	lexem, literal = "", ""
	for {
		r, err := l.readNext()
		if err != nil {
			if err == io.EOF {
				if literal == "/" {
					token = QUO
				}
				return
			}
		}
		if literal == "/" && (r == '/' || r == '*') {
			comment = literal + string(r)
			token = COMMENT
			lexem = ""
			literal += string(r)
			continue
		} else if literal == "/" {
			token = QUO
			l.backup()
			return
		}
		if r == '\n' && comment == "//" {
			token = COMMENT
			l.backup()
			return
		} else if r == '/' && len(literal) > 0 && literal[len(literal)-1] == '*' && comment == "/*" {
			literal += string(r)
			lexem = lexem[:len(lexem)-1]
			token = COMMENT
			return
		}
		literal += string(r)
		lexem += string(r)
	}
}

func (l *Lexer) lexChar() (token Token, lexem string, literal string) {
	lexem, literal = "", ""
	for {
		r, err := l.readNext()
		if err != nil {
			if err == io.EOF {
				return
			}
		}
		if IsLetter(r) && len(lexem) == 0 {
			lexem = string(r)
			literal += lexem
			continue
		}
		if r == '\'' {
			literal += string(r)
			if lexem != "" {
				token = CHAR
				return
			}
		} else {
			if len(lexem) > 0 {
				literal = literal[:1]
				lexem = literal
				l.backup()
			}
			l.backup()
			token = ILLEGAL
			return
		}
	}

}
func RuneToInt(r rune) int64 {
	if IsDigit(r) {
		return int64(r - '0')
	} else if r >= 'a' && r <= 'e' {
		return 10 + int64(r-'a')
	} else if r >= 'A' && r <= 'F' {
		return 10 + int64(r-'A')
	}
	panic("invalid hexadeciamal digit")
}

func intToString(num int64) string {
	result := ""
	if num == 0 {
		return "0"
	}
	for {
		if num == 0 {
			break
		}
		result = string(rune('0'+(num%10))) + result
		num = num / 10
	}
	return result
}

func RuneInBase(base int64, r rune) bool {
	return (base == 2 && IsBinary(r)) || (base == 8 && IsOctal(r)) || (base == 10 && IsDigit(r)) || (base == 16 && IsHex(r))
}

func IsDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func IsHex(r rune) bool {
	return IsDigit(r) || (r >= 'a' && r <= 'e') || (r >= 'A' && r <= 'E')
}

func IsLetter(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

func IsBinary(r rune) bool {
	return r == '0' || r == '1'
}

func IsOctal(r rune) bool {
	return r >= '0' && r <= '7'
}
