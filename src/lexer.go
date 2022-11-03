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
	initKeywords()
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
			return l.position, EOF, "", ""
		}
		switch r {
		case '\n':
			l.nextLine()
		case '(':
			return l.position, LPAREN, "(", string(r)
		case ')':
			return l.position, RPAREN, ")", string(r)
		case '[':
			return l.position, LBRACK, "[", string(r)
		case ']':
			return l.position, RBRACK, "]", string(r)
		case '{':
			return l.position, LBRACE, "{", string(r)
		case '}':
			return l.position, RBRACE, "}", string(r)
		case ',':
			return l.position, COMMA, ",", string(r)
		case '&':
			startPos := l.position
			token, lex, lit := l.lexAmpersand()
			return startPos, token, lex, lit
		case '|':
			startPos := l.position
			token, lex, lit := l.lexColon()
			return startPos, token, lex, lit
		case '^':
			startPos := l.position
			token, lex, lit := l.lexXor()
			return startPos, token, lex, lit
		case '%':
			startPos := l.position
			token, lex, lit := l.lexRem()
			return startPos, token, lex, lit
		case '!':
			startPos := l.position
			token, lex, lit := l.lexNot()
			return startPos, token, lex, lit
		case '.':
			startPos := l.position
			token, lex, lit := l.lexEllipsis()
			return startPos, token, lex, lit
		case '>':
			startPos := l.position
			token, lex, lit := l.lexArrowR()
			return startPos, token, lex, lit
		case '<':
			startPos := l.position
			token, lex, lit := l.lexArrowL()
			return startPos, token, lex, lit
		case ':':
			startPos := l.position
			r2, err := l.readNext()
			if err == nil && r2 == '=' {
				return startPos, DEFINE, ":=", ":="
			}
			return startPos, COLON, ":", string(r)
		case ';':
			return l.position, SEMICOLON, ";", string(r)
		case '+':
			startPos := l.position
			token, lex, lit := l.lexPlus()
			return startPos, token, lex, lit
		case '-':
			startPos := l.position
			token, lex, lit := l.lexMinus()
			return startPos, token, lex, lit
		case '*':
			startPos := l.position
			r2, err := l.readNext()
			if err == nil && r2 == '=' {
				return startPos, MUL_ASSIGN, "*=", "*="
			}
			return startPos, MUL, "*", string(r)
		case '/':
			startPos := l.position
			l.backup()
			token, lex, lit := l.lexComment()
			return startPos, token, lex, lit
		case '=':
			startPos := l.position
			r2, err := l.readNext()
			if err == nil && r2 == '=' {
				return startPos, EQL, "==", "=="
			}
			return startPos, ASSIGN, "=", string(r)
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
				keyword, ok := keywords[lex]
				if ok {
					return startPos, keyword, lex, lex
				}
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
		if err == io.EOF {
			return 0, err
		} else {
			panic(err)
		}
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
			return intToString(lexem), literal
		}
		if literal == "0_" && (r == 'O' || r == 'o' || r == 'x' || r == 'X' || r == 'b' || r == 'B') {
			panic("_ must separate successive digits")
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
		if r == '_' {
			literal += string(r)
			continue
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
			return literal
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
			if literal == "/" {
				token = QUO
			}
			return
		}
		if literal == "/" && (r == '/' || r == '*') {
			comment = literal + string(r)
			token = COMMENT
			lexem = ""
			literal += string(r)
			continue
		} else if r == '=' {
			return QUO_ASSIGN, "/=", "/="
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

func (l *Lexer) lexPlus() (Token, string, string) {
	r, err := l.readNext()
	if err != nil {
		return ADD, "+", "+"
	}
	if r == '+' {
		return INC, "++", "++"
	}
	if r == '=' {
		return ADD_ASSIGN, "+=", "+="
	}
	l.backup()
	return ADD, "+", "+"
}

func (l *Lexer) lexMinus() (Token, string, string) {
	r, err := l.readNext()
	if err != nil {
		return SUB, "-", "-"
	}
	if r == '-' {
		return DEC, "--", "--"
	}
	if r == '=' {
		return SUB_ASSIGN, "-=", "-="
	}
	l.backup()
	return SUB, "-", "-"
}

func (l *Lexer) lexEllipsis() (Token, string, string) {
	count := 1
	for {
		r, err := l.readNext()
		if err != nil {
			break
		}
		if r == '.' {
			count++
		} else {
			if count > 1 {
				l.backup()
			}
			l.backup()
			break
		}
		if count == 3 {
			return ELLIPSIS, "...", "..."
		}
	}
	return PERIOD, ".", "."
}

func (l *Lexer) lexAmpersand() (Token, string, string) {
	r, err := l.readNext()
	if err != nil {
		return AND, "&", "&"
	}
	if r == '&' {
		return LAND, "&&", "&&"
	}
	if r == '=' {
		return AND_ASSIGN, "&=", "&="
	}
	if r == '^' {
		r, err = l.readNext()
		if err == nil && r == '=' {
			return AND_NOT_ASSIGN, "&^=", "&^="
		}
		l.backup()
		return AND_NOT, "&^", "&^"
	}
	l.backup()
	return AND, "&", "&"
}

func (l *Lexer) lexColon() (Token, string, string) {
	r, err := l.readNext()
	if err != nil {
		return OR, "|", "|"
	}
	if r == '|' {
		return LOR, "||", "||"
	}
	if r == '=' {
		return OR_ASSIGN, "|=", "|="
	}
	return OR, "|", "|"
}

func (l *Lexer) lexXor() (Token, string, string) {
	r, err := l.readNext()
	if err == nil && r == '=' {
		return XOR_ASSIGN, "^=", "^="
	}
	return XOR, "^", "^"
}

func (l *Lexer) lexRem() (Token, string, string) {
	r, err := l.readNext()
	if err == nil && r == '=' {
		return REM_ASSIGN, "%=", "%="
	}
	return REM, "%", "%"
}

func (l *Lexer) lexNot() (Token, string, string) {
	r, err := l.readNext()
	if err == nil && r == '=' {
		return NEQ, "!=", "!="
	}
	return NOT, "!", "!"
}

func (l *Lexer) lexArrowR() (Token, string, string) {
	r, err := l.readNext()
	if err != nil {
		return GTR, ">", ">"
	}
	if r == '>' {
		r, err = l.readNext()
		if err == nil && r == '=' {
			return SHR_ASSIGN, ">>=", ">>="
		}
		return SHR, ">>", ">>"
	} else if r == '=' {
		return GEQ, ">=", ">="
	}
	return GTR, ">", ">"
}

func (l *Lexer) lexArrowL() (Token, string, string) {
	r, err := l.readNext()
	if err != nil {
		return LSS, "<", "<"
	}
	if r == '<' {
		r, err = l.readNext()
		if err == nil && r == '=' {
			return SHL_ASSIGN, "<<=", "<<="
		}
		return SHL, "<<", "<<"
	} else if r == '-' {
		return ARROW, "<-", "<-"
	} else if r == '=' {
		return LEQ, "<=", "<="
	}
	return LSS, "<", "<"
}

func (l *Lexer) lexChar() (token Token, lexem string, literal string) {
	lexem, literal = "", ""
	for {
		r, err := l.readNext()
		if err != nil {
			return
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
	} else if r >= 'a' && r <= 'f' {
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
	return IsDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
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
