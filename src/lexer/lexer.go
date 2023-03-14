package lexer

import (
	"bufio"
	"gocompiler/src/strconv"
	"gocompiler/src/tokens"
	"io"
	"strings"
	"unicode"
)

type Lexer struct {
	position tokens.Position
	reader   *bufio.Reader
	buffer   *Buffer
}

func NewLexer(reader io.Reader) *Lexer {
	tokens.InitKeywords()
	return &Lexer{
		position: tokens.Position{Line: 1, Column: 0},
		reader:   bufio.NewReader(reader),
		buffer:   NewBuffer(10),
	}
}

func (l *Lexer) Lex() (tokens.Position, tokens.TokenType, any, string) {
	for {
		r, err := l.readNext()
		if err != nil {
			return l.position, tokens.EOF, "", ""
		}
		switch r {
		case '\n':
			l.nextLine()
		case '(':
			return l.position, tokens.LPAREN, "(", string(r)
		case ')':
			return l.position, tokens.RPAREN, ")", string(r)
		case '[':
			return l.position, tokens.LBRACK, "[", string(r)
		case ']':
			return l.position, tokens.RBRACK, "]", string(r)
		case '{':
			return l.position, tokens.LBRACE, "{", string(r)
		case '}':
			return l.position, tokens.RBRACE, "}", string(r)
		case ',':
			return l.position, tokens.COMMA, ",", string(r)
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
				return startPos, tokens.DEFINE, ":=", ":="
			}
			return startPos, tokens.COLON, ":", string(r)
		case ';':
			return l.position, tokens.SEMICOLON, ";", string(r)
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
				return startPos, tokens.MUL_ASSIGN, "*=", "*="
			}
			l.backup()
			return startPos, tokens.MUL, "*", string(r)
		case '/':
			startPos := l.position
			l.backup()
			token, lex, lit := l.lexComment()
			return startPos, token, lex, lit
		case '=':
			startPos := l.position
			r2, err := l.readNext()
			if err == nil && r2 == '=' {
				return startPos, tokens.EQL, "==", "=="
			}
			return startPos, tokens.ASSIGN, "=", string(r)
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
				keyword, ok := tokens.Keywords[lex]
				if ok {
					return startPos, keyword, lex, lex
				}
				return startPos, tokens.IDENT, lex, lex
			} else {
				return l.position, tokens.ILLEGAL, "illegal symbol" + string(r) + "spotted", ""
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

func (l *Lexer) lexDecimal() (tokens.TokenType, any, string) {
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
			return tokens.ILLEGAL, "Unexpected error", ""
		}
		str_convert := str.String()
		if str_convert == "0_" && (r == 'O' || r == 'o' || r == 'x' || r == 'X' || r == 'b' || r == 'B') {
			return tokens.ILLEGAL, "illegal: _ must separate successive digits", ""
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
			return tokens.ILLEGAL, "illegal: hexadecimal mantissa requires p exponent", ""
		} else if (r == 'e' || r == 'E' || r == 'p' || r == 'P') && !sawexp {
			if base == 10 && (r == 'p' || r == 'P') {
				return tokens.ILLEGAL, "illegal: p exponent requires hexadecimal mantissa", ""
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
		return tokens.ILLEGAL, "Unexpected error", ""
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
			return tokens.ILLEGAL, "Unexpected error", ""
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
			var d strconv.Decimal
			err := d.FromString(str_convert)
			if err != nil {
				return tokens.ILLEGAL, err.Error(), ""
			}
			bits, _ := d.FloatBits()
			return tokens.FLOAT, strconv.Float32FromBits(uint32(bits)), literal.String()
		case 16:
			lex, err := strconv.BitsFromHex(str.String(), mantiss, exponent, truncate)
			if err != nil {
				return tokens.ILLEGAL, "unexpeted error", ""
			}
			return tokens.FLOAT, float32(lex), literal.String() // reading Hex
		default:
			return tokens.ILLEGAL, "illegal: p exponent requires hexadecimal mantissa", ""
		}
	} else {
		return l.lexInt(str.String())
	}
}

func (l *Lexer) lexInt(s string) (tokens.TokenType, any, string) {
	var lexem, base, additional_lexem int64 = 0, 10, 0
	var literal strings.Builder
	uncertainBase := false
	for _, r := range s {
		str_convert := literal.String()
		if str_convert == "0_" && (r == 'O' || r == 'o' || r == 'x' || r == 'X' || r == 'b' || r == 'B') {
			return tokens.ILLEGAL, "illegal: _ must separate successive digits", ""
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
			return tokens.IMAG, intToString(lexem) + "i", literal.String() //TODO
		}
		if RuneInBase(base, r) {
			literal.WriteRune(r)
			if uncertainBase {
				additional_lexem = RuneToInt(r) + additional_lexem*10
			}
			lexem = RuneToInt(r) + lexem*base
			if lexem > 2147483647 {
				return tokens.ILLEGAL, "illegal: integer value overflow", ""
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
	return tokens.INT, int32(lexem), literal.String()
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

func (l *Lexer) lexComment() (token tokens.TokenType, lexem string, literal string) {
	comment := ""
	lexem, literal = "", ""
	for {
		r, err := l.readNext()
		if err != nil {
			if literal == "/" {
				token = tokens.QUO
			}
			return
		}
		if literal == "/" {
			if r == '/' || r == '*' {
				comment = literal + string(r)
				token = tokens.COMMENT
				lexem = ""
				literal += string(r)
				continue
			} else if r == '=' {
				return tokens.QUO_ASSIGN, "/=", "/="
			} else if literal == "/" {
				token = tokens.QUO
				l.backup()
				return
			}
		}
		if r == '\n' {
			if comment == "//" {
				token = tokens.COMMENT
				l.backup()
				return
			}
			l.nextLine()
		} else if r == '/' && len(literal) > 0 && literal[len(literal)-1] == '*' && comment == "/*" {
			literal += string(r)
			lexem = lexem[:len(lexem)-1]
			token = tokens.COMMENT
			return
		}
		literal += string(r)
		lexem += string(r)
	}
}

func (l *Lexer) lexPlus() (tokens.TokenType, string, string) {
	r, err := l.readNext()
	if err != nil {
		return tokens.ADD, "+", "+"
	}
	if r == '+' {
		return tokens.INC, "++", "++"
	}
	if r == '=' {
		return tokens.ADD_ASSIGN, "+=", "+="
	}
	l.backup()
	return tokens.ADD, "+", "+"
}

func (l *Lexer) lexMinus() (tokens.TokenType, string, string) {
	r, err := l.readNext()
	if err != nil {
		return tokens.SUB, "-", "-"
	}
	if r == '-' {
		return tokens.DEC, "--", "--"
	}
	if r == '=' {
		return tokens.SUB_ASSIGN, "-=", "-="
	}
	l.backup()
	return tokens.SUB, "-", "-"
}

func (l *Lexer) lexEllipsis() (tokens.TokenType, any, string) {
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
			return tokens.ELLIPSIS, "...", "..."
		}
	}
	return tokens.PERIOD, ".", "."
}

func (l *Lexer) lexAmpersand() (tokens.TokenType, string, string) {
	r, err := l.readNext()
	if err != nil {
		return tokens.AND, "&", "&"
	}
	if r == '&' {
		return tokens.LAND, "&&", "&&"
	}
	if r == '=' {
		return tokens.AND_ASSIGN, "&=", "&="
	}
	if r == '^' {
		r, err = l.readNext()
		if err == nil && r == '=' {
			return tokens.AND_NOT_ASSIGN, "&^=", "&^="
		}
		l.backup()
		return tokens.AND_NOT, "&^", "&^"
	}
	l.backup()
	return tokens.AND, "&", "&"
}

func (l *Lexer) lexColon() (tokens.TokenType, string, string) {
	r, err := l.readNext()
	if err != nil {
		return tokens.OR, "|", "|"
	}
	if r == '|' {
		return tokens.LOR, "||", "||"
	}
	if r == '=' {
		return tokens.OR_ASSIGN, "|=", "|="
	}
	return tokens.OR, "|", "|"
}

func (l *Lexer) lexXor() (tokens.TokenType, string, string) {
	r, err := l.readNext()
	if err == nil && r == '=' {
		return tokens.XOR_ASSIGN, "^=", "^="
	}
	return tokens.XOR, "^", "^"
}

func (l *Lexer) lexRem() (tokens.TokenType, string, string) {
	r, err := l.readNext()
	if err == nil && r == '=' {
		return tokens.REM_ASSIGN, "%=", "%="
	}
	return tokens.REM, "%", "%"
}

func (l *Lexer) lexNot() (tokens.TokenType, string, string) {
	r, err := l.readNext()
	if err == nil && r == '=' {
		return tokens.NEQ, "!=", "!="
	}
	return tokens.NOT, "!", "!"
}

func (l *Lexer) lexArrowR() (tokens.TokenType, string, string) {
	r, err := l.readNext()
	if err != nil {
		return tokens.GTR, ">", ">"
	}
	if r == '>' {
		r, err = l.readNext()
		if err == nil && r == '=' {
			return tokens.SHR_ASSIGN, ">>=", ">>="
		}
		return tokens.SHR, ">>", ">>"
	} else if r == '=' {
		return tokens.GEQ, ">=", ">="
	}
	return tokens.GTR, ">", ">"
}

func (l *Lexer) lexArrowL() (tokens.TokenType, string, string) {
	r, err := l.readNext()
	if err != nil {
		return tokens.LSS, "<", "<"
	}
	if r == '<' {
		r, err = l.readNext()
		if err == nil && r == '=' {
			return tokens.SHL_ASSIGN, "<<=", "<<="
		}
		return tokens.SHL, "<<", "<<"
	} else if r == '-' {
		return tokens.ARROW, "<-", "<-"
	} else if r == '=' {
		return tokens.LEQ, "<=", "<="
	}
	return tokens.LSS, "<", "<"
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

func (l *Lexer) lexChar() (token tokens.TokenType, lexem string, literal string) {
	lexem, literal, error := l.lexCharSymbol()
	if len(error) > 0 {
		return tokens.ILLEGAL, error, ""
	}
	literal = "'" + literal
	r, err := l.readNext()
	if err != nil {
		return
	}
	if r == '\'' && len(lexem) > 0 {
		literal += string(r)
		return tokens.CHAR, lexem, literal
	}
	for i := 0; i < len(literal)-1; i++ {
		l.backup()
	}
	return tokens.ILLEGAL, "illegal: rune literal not terminated", ""
}

func (l *Lexer) lexString(isRaw bool) (token tokens.TokenType, lexem string, literal string) {
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
				l.nextLine()
				literal += string(r)
				lexem += string(r)
				continue
			}
			return tokens.ILLEGAL, "illegal: string literal not terminated", ""
		}
		l.backup()
		char_lexem, char_literal, error := l.lexCharSymbol()
		if len(error) > 0 {
			return tokens.ILLEGAL, error, ""
		}
		lexem += char_lexem
		literal += char_literal
	}
	if isRaw {
		literal = "`" + literal
	} else {
		literal = "\"" + literal
	}
	return tokens.STRING, lexem, literal
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
