package lexer

import (
	"bufio"
	"io"
	"strings"
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
		buffer:   NewBuffer(10),
	}
}

func (l *Lexer) Lex() (Position, Token, any, string) {
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
		case '"', '`':
			startPos := l.position
			token, lex, lit := l.lexString(r == '`')
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
			token, lex, lit := l.lexChar()
			return startPos, token, lex, lit
		default:
			if unicode.IsSpace(r) {
				continue
			} else if IsDigit(r) {
				startPos := l.position
				l.backup()
				token, lex, lit := l.lexDecimal()
				return startPos, token, lex, lit
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
				return l.position, ILLEGAL, "illegal symbol" + string(r) + "spotted", ""
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

func (l *Lexer) lexDecimal() (Token, any, string) {
	var str strings.Builder
	var literal strings.Builder
	var r rune
	var err error
	var base uint64 = 10
	var sawdot = false
	var sawexp = false

	var maxMantiss = 19

	ndigits := 0
	ndMant := 0
	var mantiss uint64
	var truncate = false
	var exponent int
	var pointIndex int
	var esign int = 1
	var e int = 0

	for i := 0; ; i++ {
		r, err = l.readNext()
		if err != nil {
			if err == io.EOF {
				goto out
			}
			return ILLEGAL, "Unexpected error", ""
		}
		str_convert := str.String()
		if str_convert == "0_" && (r == 'O' || r == 'o' || r == 'x' || r == 'X' || r == 'b' || r == 'B') {
			return ILLEGAL, "illegal: _ must separate successive digits", ""
		}
		if str_convert == "0" {
			switch r {
			case 'x', 'X':
				base = 16
				maxMantiss = 16
				literal.WriteRune(r)
				str.WriteRune(r)
				continue
			case 'o', 'O':
				base = 8
				literal.WriteRune(r)
				str.WriteRune(r)
				continue
			case 'b', 'B':
				base = 2
				literal.WriteRune(r)
				str.WriteRune(r)
				continue
			}
		}
		if r == '_' {
			str.WriteRune(r)
			literal.WriteRune(r)
			continue
		}

		if r == '.' && !sawdot {
			sawdot = true
			pointIndex = ndigits
			literal.WriteRune(r)
			str.WriteRune(r)
			continue
		}
		if RuneInBase(int64(base), r) {
			str.WriteRune(r)
			literal.WriteRune(r)
			if r == '0' && ndigits == 0 {
				pointIndex--
				continue
			}
			ndigits++
			if ndMant < maxMantiss {
				mantiss *= base
				mantiss += uint64(RuneToInt(r))
				ndMant++
			} else {
				truncate = true
			}
		} else if (r == '-' || r == '+') && base == 16 && (str_convert[len(str_convert)-1] == 'e' || str_convert[len(str_convert)-1] == 'E') && sawdot {
			return ILLEGAL, "illegal: hexadecimal mantissa requires p exponent", ""
		} else if (r == 'e' || r == 'E' || r == 'p' || r == 'P') && !sawexp {
			if base == 10 && (r == 'p' || r == 'P') {
				return ILLEGAL, "illegal: p exponent requires hexadecimal mantissa", ""
			}
			literal.WriteRune(r)
			str.WriteRune(r)
			sawexp = true
			goto loop
		} else {
			l.backup()
			goto out
		}
	}
loop:
	if !sawdot {
		pointIndex = ndigits
	}
	if base == 16 {
		pointIndex *= 4
		ndMant *= 4
	}
	r, err = l.readNext()
	if err != nil {
		return ILLEGAL, "Unexpected error", ""
	}
	if r == '-' || r == '+' {
		literal.WriteRune(r)
		str.WriteRune(r)
		if r == '-' {
			esign = -1
		}
	} else {
		l.backup()
	}
	for r, err = l.readNext(); IsDigit(r) || r == '_'; r, err = l.readNext() {
		if err != nil {
			if err == io.EOF {
				break
			}
			return ILLEGAL, "Unexpected error", ""
		}
		literal.WriteRune(r)
		if r == '_' {
			continue
		}
		str.WriteRune(r)
		if e < 10000 {
			e = e*10 + int(RuneToInt(r))
		}
	}
	pointIndex += e * esign
	if mantiss != 0 {
		exponent = pointIndex - ndMant
	}
out:
	if sawdot || sawexp {
		if str.String() == "." {
			l.backup()
			return l.lexEllipsis()
		}
		switch base {
		case 10:
			str_convert := str.String()
			var d decimal
			err := d.fromString(str_convert)
			if err != nil {
				return ILLEGAL, err.Error(), ""
			}
			bits, _ := d.floatBits()
			return FLOAT, Float32FromBits(uint32(bits)), literal.String()
		case 16:
			lex, err := bitsFromHex(str.String(), mantiss, exponent, truncate)
			if err != nil {
				return ILLEGAL, "unexpeted error", ""
			}
			return FLOAT, float32(lex), literal.String() // reading Hex
		default:
			return ILLEGAL, "illegal: p exponent requires hexadecimal mantissa", ""
		}
	} else {
		return l.lexInt(str.String())
	}
}

func (l *Lexer) lexInt(s string) (Token, any, string) {
	var lexem, base, additional_lexem int64 = 0, 10, 0
	var literal strings.Builder
	uncertainBase := false
	for _, r := range s {
		str_convert := literal.String()
		if str_convert == "0_" && (r == 'O' || r == 'o' || r == 'x' || r == 'X' || r == 'b' || r == 'B') {
			return ILLEGAL, "illegal: _ must separate successive digits", ""
		}
		if str_convert == "0" {
			switch r {
			case 'x', 'X':
				base = 16
				literal.WriteRune(r)
				continue
			case 'o', 'O':
				base = 8
				literal.WriteRune(r)
				continue
			case 'b', 'B':
				base = 2
				literal.WriteRune(r)
				continue
			default:
				base = 8
				uncertainBase = true
			}
		}
		if r == '_' {
			literal.WriteRune(r)
			continue
		}
		if (r == '.' || r == 'p' || r == 'P') && (base == 10 || base == 16 || uncertainBase) || ((r == 'e' || r == 'E') && (base == 10 || uncertainBase)) {
			if r == 'p' || r == 'P' || r == 'e' || r == 'E' {
				l.backup()
			} else {
				literal.WriteRune(r)
			}
		}
		if r == 'i' {
			literal.WriteRune(r)
			return IMAG, intToString(lexem) + "i", literal.String() //TODO
		}
		if RuneInBase(base, r) {
			literal.WriteRune(r)
			if uncertainBase {
				additional_lexem = RuneToInt(r) + additional_lexem*10
			}
			lexem = RuneToInt(r) + lexem*base
			if lexem > 2147483647 {
				return ILLEGAL, "illegal: integer value overflow", ""
			}
		} else {
			break
		}
	}
	if literal.String() == "0x" {
		literal.Reset()
		literal.WriteString("0")
		l.backup()
	}
	return INT, int32(lexem), literal.String()
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
		if literal == "/" {
			if r == '/' || r == '*' {
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

func (l *Lexer) lexEllipsis() (Token, any, string) {
	count := 1
	for {
		r, err := l.readNext()
		if err != nil {
			break
		}
		if IsDigit(r) {
			l.backup()
			l.backup()
			return l.lexDecimal()
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

func (l *Lexer) lexCharSymbol() (lexem string, literal string, error string) {
	lexem, literal = "", ""
	var i int
	byte_base, byte_limit, byte_code := 0, 1, 0
	for i = 1; i <= byte_limit; i++ {
		r, err := l.readNext()
		if err != nil {
			return
		}
		if r == '\n' {
			l.backup()
			return "", "", ""
		}
		if byte_base == 8 {
			if r == 'u' {
				literal += string(r)
				byte_base = 16
				byte_limit = 4
				i = 0
				continue
			} else if r == 'U' {
				literal += string(r)
				byte_base = 16
				byte_limit = 8
				i = 0
				continue
			} else if r == 'x' {
				literal += string(r)
				byte_base = 16
				byte_limit = 2
				i = 0
				continue
			}
			val, ok := isEscapedChar(r)
			if ok {
				literal += string(r)
				lexem = string(val)
				byte_base = 0
				i = 0
				break
			}
			if IsOctal(r) {
				lexem += string(r)
				literal += string(r)
				byte_code = int(RuneToInt(r)) + byte_code*byte_base
				continue
			}
			l.backup()
			break
		}
		if byte_base > 0 {
			if RuneInBase(int64(byte_base), r) {
				lexem += string(r)
				literal += string(r)
				byte_code = int(RuneToInt(r)) + byte_code*byte_base
				continue
			}
			l.backup()
			break

		} else if r == '\\' {
			literal += string(r)
			byte_base = 8
			byte_limit = 3
			i = 0
			continue
		}
		lexem += string(r)
		literal += string(r)
	}
	if byte_base > 0 {
		if byte_base == 8 && (byte_code < 0 || byte_code > 255) {
			return "", "", "illegal: octal value over 255"
		}
		if byte_base == 16 && (byte_code > 0x10FFFF || (byte_code >= 0xD800 && byte_code <= 0xDFFF)) {
			return "", "", "illegal: invalid Unicode code point"
		}
		lexem = string(rune(byte_code))
	} else if i > 2 {
		return "", "", "illegal: too many characters"
	}
	return
}

func (l *Lexer) lexChar() (token Token, lexem string, literal string) {
	lexem, literal, error := l.lexCharSymbol()
	if len(error) > 0 {
		return ILLEGAL, error, ""
	}
	literal = "'" + literal
	r, err := l.readNext()
	if err != nil {
		return
	}
	if r == '\'' && len(lexem) > 0 {
		literal += string(r)
		return CHAR, lexem, literal
	}
	for i := 0; i < len(literal)-1; i++ {
		l.backup()
	}
	return ILLEGAL, "illegal: rune literal not terminated", ""
}

func (l *Lexer) lexString(isRaw bool) (token Token, lexem string, literal string) {
	for {
		r, err := l.readNext()
		if err != nil {
			return
		}
		if (isRaw && r == '`') || (!isRaw && r == '"') {
			literal += string(r)
			break
		}
		if r == '\n' {
			if isRaw {
				literal += string(r)
				lexem += string(r)
				continue
			}
			return ILLEGAL, "illegal: string literal not terminated", ""
		}
		l.backup()
		char_lexem, char_literal, error := l.lexCharSymbol()
		if len(error) > 0 {
			return ILLEGAL, error, ""
		}
		lexem += char_lexem
		literal += char_literal
	}
	if isRaw {
		literal = "`" + literal
	} else {
		literal = "\"" + literal
	}
	return STRING, lexem, literal
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

func isEscapedChar(r rune) (rune, bool) {
	var chars = map[rune]rune{'a': '\a', 'b': '\b', 'f': '\f', 'n': '\n', 'r': '\r', 't': '\t', 'v': '\v', '\\': '\\', '\'': '\'', '"': '"'}
	symbol, ok := chars[r]
	return symbol, ok
}
