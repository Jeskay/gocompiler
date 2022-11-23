package lexer

import (
	"fmt"
	"strings"
	"testing"
)

type lexem struct {
	pos Position
	tok Token
	lex string
	lit string
}

func (l *lexem) Compare(l2 lexem) bool {
	return l2.pos.Line == l.pos.Line && l2.pos.Column == l.pos.Column && l2.tok == l.tok && l2.lit == l.lit && l.lex == l2.lex
}
func (l *lexem) ToString() string {
	return fmt.Sprintf("%d:%d\t%s\t%s\t%s\n", l.pos.Line, l.pos.Column, l.tok, l.lex, l.lit)
}
func performTest(t *testing.T, input string, expect []lexem) {
	lexerInstance := NewLexer(strings.NewReader(input))
	for i := 0; ; i++ {
		pos, tok, lex, lit := lexerInstance.Lex()
		if tok == EOF {
			break
		}
		got := lexem{pos, tok, lex, lit}
		if !got.Compare(expect[i]) {
			t.Errorf("expected %s, got %s", expect[i].ToString(), got.ToString())
		}
		if tok == ILLEGAL {
			break
		}
	}
}
func TestIntDigits(t *testing.T) {
	var expected = [...]lexem{
		{Position{1, 1}, INT, "1910", "1910"},
		{Position{2, 1}, INT, "0", "0"},
		{Position{3, 1}, INT, "4", "0b100"},
		{Position{4, 1}, INT, "7", "0b00111"},
		{Position{5, 1}, INT, "511", "0777"},
		{Position{6, 1}, INT, "668", "0o1234"},
		{Position{7, 1}, INT, "282", "0O0432"},
		{Position{8, 1}, INT, "427", "0x01AB"},
		{Position{9, 1}, INT, "171", "0Xab"},
		{Position{10, 1}, INT, "384", "0_600"},
		{Position{11, 1}, INT, "195951310", "0xBadFace"},
		{Position{12, 1}, INT, "195951310", "0xBad_Face"},
	}
	const input = "1910 \n0 \n0b100 \n0b00111 \n0777 \n0o1234 \n0O0432 \n0x01AB \n0Xab \n0_600 \n0xBadFace \n0xBad_Face"
	performTest(t, input, expected[:])
	performTest(t, "2147483649", []lexem{{Position{1, 1}, ILLEGAL, "illegal: integer value overflow", ""}})
	performTest(t, "0_xBadFace", []lexem{{Position{1, 1}, ILLEGAL, "illegal: _ must separate successive digits", ""}})
}

func TestFloatDigits(t *testing.T) {
	expected := [...]lexem{
		{Position{1, 1}, FLOAT, "15", "0.15e+0_2"},
		{Position{2, 1}, FLOAT, "2048", "0x2.p10"},
		{Position{3, 1}, FLOAT, "2.71828", "2.71828"},
		{Position{4, 1}, FLOAT, "0", "0000."},
		{Position{5, 1}, FLOAT, "72.4", "072.40"},
		{Position{6, 1}, FLOAT, "1.9375", "0x1.Fp+0"},
		{Position{7, 1}, FLOAT, "1", "1.e+0"},
		{Position{8, 1}, FLOAT, "0.0000000000667428", "6.67428e-11"},
		{Position{9, 1}, FLOAT, "1000000", "1E6"},
		{Position{10, 1}, FLOAT, "0.25", ".25"},
		{Position{11, 1}, FLOAT, "12345", ".12345E+5"},
		{Position{12, 1}, FLOAT, "15", "1_5."},
		{Position{13, 1}, FLOAT, "15", "0.15e+0_2"},
		{Position{14, 1}, FLOAT, "0.25", "0x1p-2"},
		{Position{15, 1}, FLOAT, "2048", "0x2.p10"},
		{Position{16, 1}, FLOAT, "1.9375", "0x1.Fp+0"},
		{Position{17, 1}, FLOAT, "0.5", "0X.8p-0"},
		{Position{18, 1}, FLOAT, "0.1249847412109375", "0X_1FFFP-16"},
		{Position{19, 1}, INT, "350", "0x15e"},
		{Position{19, 6}, SUB, "-", "-"},
		{Position{19, 7}, INT, "2", "2"},
	}
	const input = "0.15e+0_2 \n0x2.p10 \n2.71828 \n0000. \n072.40 \n0x1.Fp+0 \n1.e+0 \n6.67428e-11 \n1E6 \n.25 \n.12345E+5 \n1_5. \n0.15e+0_2  \n0x1p-2 \n0x2.p10 \n0x1.Fp+0 \n0X.8p-0 \n0X_1FFFP-16 \n0x15e-2"
	performTest(t, input, expected[:])
	performTest(t, "1p-2", []lexem{{Position{1, 1}, ILLEGAL, "illegal: p exponent requires hexadecimal mantissa", ""}})
	performTest(t, "0x1.5e-2", []lexem{{Position{1, 1}, ILLEGAL, "illegal: hexadecimal mantissa requires p exponent", ""}})
}
func TestIdents(t *testing.T) {
	expected := [...]lexem{
		{Position{1, 1}, IDENT, "test", "test"},
		{Position{2, 1}, INT, "2", "2"},
		{Position{2, 2}, IDENT, "test", "test"},
		{Position{3, 1}, IDENT, "test2", "test2"},
		{Position{4, 1}, IDENT, "Test2", "Test2"},
		{Position{5, 1}, IDENT, "TEST", "TEST"},
		{Position{6, 1}, IDENT, "_test2", "_test2"},
		{Position{7, 1}, IDENT, "_2test", "_2test"},
		{Position{8, 1}, IDENT, "___", "___"},
		{Position{9, 2}, IDENT, "αβ", "αβ"},
	}
	const input = "test \n2test \ntest2 \nTest2 \nTEST \n_test2 \n_2test \n___ \n αβ"
	performTest(t, input, expected[:])
}

func TestComment(t *testing.T) {
	expected := [...]lexem{
		{Position{1, 1}, INT, "1", "1"},
		{Position{1, 3}, QUO, "/", "/"},
		{Position{1, 5}, INT, "2", "2"},
		{Position{2, 1}, COMMENT, "комментарий 1", "//комментарий 1"},
		{Position{3, 1}, COMMENT, " 1 + 5 / 10", "// 1 + 5 / 10"},
		{Position{4, 1}, COMMENT, "comment 1", "/*comment 1*/"},
		{Position{5, 1}, COMMENT, "/sdfsdf", "///sdfsdf"},
		{Position{6, 1}, COMMENT, "1t\n2t\n3t", "/*1t\n2t\n3t*/"},
		{Position{7, 1}, IDENT, "end", "end"},
		{Position{8, 1}, COMMENT, "comment with no ending quote", "/*comment with no ending quote"},
	}
	const input = "1 / 2\n//комментарий 1\n// 1 + 5 / 10\n/*comment 1*/\n///sdfsdf\n/*1t\n2t\n3t*/\nend\n/*comment with no ending quote"
	performTest(t, input, expected[:])
}

func TestOperands(t *testing.T) {
	expected := [...]lexem{
		{Position{1, 1}, ADD, "+", "+"},
		{Position{1, 6}, AND, "&", "&"},
		{Position{1, 12}, ADD_ASSIGN, "+=", "+="},
		{Position{1, 18}, AND_ASSIGN, "&=", "&="},
		{Position{1, 25}, LAND, "&&", "&&"},
		{Position{1, 31}, EQL, "==", "=="},
		{Position{1, 37}, NEQ, "!=", "!="},
		{Position{1, 43}, LPAREN, "(", "("},
		{Position{1, 48}, RPAREN, ")", ")"},
		{Position{2, 1}, SUB, "-", "-"},
		{Position{2, 6}, OR, "|", "|"},
		{Position{2, 12}, SUB_ASSIGN, "-=", "-="},
		{Position{2, 18}, OR_ASSIGN, "|=", "|="},
		{Position{2, 25}, LOR, "||", "||"},
		{Position{2, 31}, LSS, "<", "<"},
		{Position{2, 37}, LEQ, "<=", "<="},
		{Position{2, 43}, LBRACK, "[", "["},
		{Position{2, 48}, RBRACK, "]", "]"},
		{Position{3, 1}, MUL, "*", "*"},
		{Position{3, 6}, XOR, "^", "^"},
		{Position{3, 12}, MUL_ASSIGN, "*=", "*="},
		{Position{3, 18}, XOR_ASSIGN, "^=", "^="},
		{Position{3, 25}, ARROW, "<-", "<-"},
		{Position{3, 31}, GTR, ">", ">"},
		{Position{3, 37}, GEQ, ">=", ">="},
		{Position{3, 43}, LBRACE, "{", "{"},
		{Position{3, 48}, RBRACE, "}", "}"},
		{Position{4, 1}, QUO, "/", "/"},
		{Position{4, 6}, SHL, "<<", "<<"},
		{Position{4, 12}, QUO_ASSIGN, "/=", "/="},
		{Position{4, 18}, SHL_ASSIGN, "<<=", "<<="},
		{Position{4, 25}, INC, "++", "++"},
		{Position{4, 31}, ASSIGN, "=", "="},
		{Position{4, 37}, DEFINE, ":=", ":="},
		{Position{4, 43}, COMMA, ",", ","},
		{Position{4, 48}, SEMICOLON, ";", ";"},
		{Position{5, 1}, REM, "%", "%"},
		{Position{5, 6}, SHR, ">>", ">>"},
		{Position{5, 12}, REM_ASSIGN, "%=", "%="},
		{Position{5, 18}, SHR_ASSIGN, ">>=", ">>="},
		{Position{5, 25}, DEC, "--", "--"},
		{Position{5, 31}, NOT, "!", "!"},
		{Position{5, 37}, ELLIPSIS, "...", "..."},
		{Position{5, 43}, PERIOD, ".", "."},
		{Position{5, 48}, COLON, ":", ":"},
		{Position{6, 6}, AND_NOT, "&^", "&^"},
		{Position{6, 18}, AND_NOT_ASSIGN, "&^=", "&^="},
	}
	const input = "+    &     +=    &=     &&    ==    !=    (    )\n-    |     -=    |=     ||    <     <=    [    ]\n*    ^     *=    ^=     <-    >     >=    {    }\n/    <<    /=    <<=    ++    =     :=    ,    ;\n%    >>    %=    >>=    --    !     ...   .    : \n     &^          &^= "
	performTest(t, input, expected[:])
}

func TestChar(t *testing.T) {
	expected := [...]lexem{
		{Position{1, 1}, CHAR, "a", "'a'"},
		{Position{2, 1}, CHAR, "ä", "'ä'"},
		{Position{3, 1}, CHAR, "本", "'本'"},
		{Position{4, 1}, CHAR, "\t", "'\t'"},
		{Position{5, 1}, CHAR, "\u12e4", "'\u12e4'"},
		{Position{6, 1}, CHAR, "\U00101234", "'\U00101234'"},
	}
	const input = "'a' \n'ä' \n'本' \n'\t' \n'\u12e4' \n'\U00101234'"
	performTest(t, input, expected[:])
	performTest(t, "'aa'", []lexem{{Position{1, 1}, ILLEGAL, "illegal: rune literal not terminated", ""}})
	performTest(t, "1p-2", []lexem{{Position{1, 1}, ILLEGAL, "illegal: p exponent requires hexadecimal mantissa", ""}})
}

func TestString(t *testing.T) {
	expected := [...]lexem{
		{Position{1, 1}, STRING, "abc", "`abc`"},
		{Position{2, 1}, STRING, "\n\n\n", "`\\n\n\\n`"},
		{Position{3, 1}, STRING, "\n", "\"\\n\""},
		{Position{4, 1}, STRING, `"`, `"\""`},
		{Position{5, 1}, STRING, "Hello, world!\n", `"Hello, world!\n"`},
		{Position{6, 1}, STRING, `日本語`, `"日本語"`},
	}
	const input = "`abc` \n`\\n\n\\n` \n\"\\n\" \n\"\\\"\" \n\"Hello, world!\\n\" \n\"日本語\""
	performTest(t, input, expected[:])
	expected2 := [...]lexem{
		{Position{1, 1}, STRING, "日本語", `"\u65e5本\U00008a9e"`},
		{Position{2, 2}, STRING, "ÿÿ", `"\xff\u00FF"`},
	}
	const input2 = `"\u65e5本\U00008a9e"
	"\xff\u00FF"`
	performTest(t, input2, expected2[:])
	performTest(t, `"\uD800"`, []lexem{{Position{1, 1}, ILLEGAL, "illegal: invalid Unicode code point", ""}})
	performTest(t, `"\U00110000"`, []lexem{{Position{1, 1}, ILLEGAL, "illegal: invalid Unicode code point", ""}})
}

func TestStringFormats(t *testing.T) {
	expected := [...]lexem{
		{Position{1, 1}, STRING, "日本語", `"日本語"`},
		{Position{1, 7}, STRING, "日本語", `"\u65e5\u672c\u8a9e"`},
		{Position{1, 28}, STRING, "日本語", `"\U000065e5\U0000672c\U00008a9e"`},
	}
	const input = `"日本語" "\u65e5\u672c\u8a9e" "\U000065e5\U0000672c\U00008a9e"`
	performTest(t, input, expected[:])
}

func TestHelloWorld(t *testing.T) {
	expected := []lexem{
		{Position{1, 1}, PACKAGE, "package", "package"},
		{Position{1, 9}, IDENT, "hello", "hello"},
		{Position{3, 1}, IMPORT, "import", "import"},
		{Position{3, 8}, LPAREN, "(", "("},
		{Position{4, 5}, STRING, "fmt", `"fmt"`},
		{Position{5, 1}, RPAREN, ")", ")"},
		{Position{6, 1}, COMMENT, "\nsimple programm that greets you!\naccepts nothing\n", "/*\nsimple programm that greets you!\naccepts nothing\n*/"},
		{Position{7, 1}, FUNC, "func", "func"},
		{Position{7, 6}, IDENT, "main", "main"},
		{Position{7, 10}, LPAREN, "(", "("},
		{Position{7, 11}, RPAREN, ")", ")"},
		{Position{7, 13}, LBRACE, "{", "{"},
		{Position{8, 5}, CONST, "const", "const"},
		{Position{8, 11}, IDENT, "message", "message"},
		{Position{8, 19}, ASSIGN, "=", "="},
		{Position{8, 21}, STRING, "Hello world!\nend of the message", "`Hello world!\nend of the message`"},
		{Position{9, 5}, IDENT, "fmt", "fmt"},
		{Position{9, 8}, PERIOD, ".", "."},
		{Position{9, 9}, IDENT, "Printf", "Printf"},
		{Position{9, 15}, LPAREN, "(", "("},
		{Position{9, 16}, IDENT, "message", "message"},
		{Position{9, 23}, RPAREN, ")", ")"},
		{Position{10, 1}, RBRACE, "}", "}"},
	}
	const input = "package hello\n\nimport (\n    \"fmt\"\n)\n/*\nsimple programm that greets you!\naccepts nothing\n*/\nfunc main() {\n    const message = `Hello world!\nend of the message`\n    fmt.Printf(message)\n}"
	performTest(t, input, expected[:])
}
