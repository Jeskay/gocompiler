package lexer

import (
	"gocompiler/src/tokens"
	"io"
	"math"
	"os"
	"strings"
	"testing"
)

const float64EqualityThreshold = 1e-9

func readInput(filename string) string {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0600)
	if err != nil {
		panic(err)
	}
	b, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	return strings.ReplaceAll(string(b), "\r", "")
}

func compareFloat32(a, b float32) bool {
	return math.Abs(float64(a-b)) <= float64EqualityThreshold
}
func CompareTokens(l tokens.Token, l2 tokens.Token) bool {
	var isLexemEqual bool
	var f1, ok1 = l.Lex.(float32)
	var f2, ok2 = l.Lex.(float32)
	if ok1 && ok2 {
		isLexemEqual = compareFloat32(f1, f2)
	} else {
		isLexemEqual = l.Lex == l2.Lex
	}
	return l2.Pos.Line == l.Pos.Line && l2.Pos.Column == l.Pos.Column && l2.Tok == l.Tok && l2.Lit == l.Lit && isLexemEqual
}

func performTest(t *testing.T, input string, expect []tokens.Token) {
	lexerInstance := NewLexer(strings.NewReader(input))
	for i := 0; ; i++ {
		pos, tok, lex, lit := lexerInstance.Lex()
		if tok == tokens.EOF {
			break
		}
		got := tokens.Token{Pos: pos, Tok: tok, Lex: lex, Lit: lit}
		if !CompareTokens(got, expect[i]) {
			t.Errorf("expected %s, got %s", expect[i].ToString(), got.ToString())
		}
		if tok == tokens.ILLEGAL {
			break
		}
	}
}
func TestIntDigits(t *testing.T) {
	var expected = [...]tokens.Token{
		{Pos: tokens.Position{Line: 1, Column: 1}, Tok: tokens.INT, Lex: int32(1910), Lit: "1910"},
		{Pos: tokens.Position{Line: 2, Column: 1}, Tok: tokens.INT, Lex: int32(0), Lit: "0"},
		{Pos: tokens.Position{Line: 3, Column: 1}, Tok: tokens.INT, Lex: int32(4), Lit: "0b100"},
		{Pos: tokens.Position{Line: 4, Column: 1}, Tok: tokens.INT, Lex: int32(7), Lit: "0b00111"},
		{Pos: tokens.Position{Line: 5, Column: 1}, Tok: tokens.INT, Lex: int32(511), Lit: "0777"},
		{Pos: tokens.Position{Line: 6, Column: 1}, Tok: tokens.INT, Lex: int32(668), Lit: "0o1234"},
		{Pos: tokens.Position{Line: 7, Column: 1}, Tok: tokens.INT, Lex: int32(282), Lit: "0O0432"},
		{Pos: tokens.Position{Line: 8, Column: 1}, Tok: tokens.INT, Lex: int32(427), Lit: "0x01AB"},
		{Pos: tokens.Position{Line: 9, Column: 1}, Tok: tokens.INT, Lex: int32(171), Lit: "0Xab"},
		{Pos: tokens.Position{Line: 10, Column: 1}, Tok: tokens.INT, Lex: int32(384), Lit: "0_600"},
		{Pos: tokens.Position{Line: 11, Column: 1}, Tok: tokens.INT, Lex: int32(195951310), Lit: "0xBadFace"},
		{Pos: tokens.Position{Line: 12, Column: 1}, Tok: tokens.INT, Lex: int32(195951310), Lit: "0xBad_Face"},
	}
	input := readInput("../tests/lexer/test1.txt")
	performTest(t, input, expected[:])
	performTest(t, "2147483649", []tokens.Token{{Pos: tokens.Position{Line: 1, Column: 1}, Tok: tokens.ILLEGAL, Lex: "illegal: integer value overflow", Lit: ""}})
	performTest(t, "0_xBadFace", []tokens.Token{{Pos: tokens.Position{Line: 1, Column: 1}, Tok: tokens.ILLEGAL, Lex: "illegal: _ must separate successive digits", Lit: ""}})
}

func TestFloatDigits(t *testing.T) {
	expected := [...]tokens.Token{
		{Pos: tokens.Position{Line: 1, Column: 1}, Tok: tokens.FLOAT, Lex: float32(0.15e+0_2), Lit: "0.15e+0_2"},
		{Pos: tokens.Position{Line: 2, Column: 1}, Tok: tokens.FLOAT, Lex: float32(0x2.p10), Lit: "0x2.p10"},
		{Pos: tokens.Position{Line: 3, Column: 1}, Tok: tokens.FLOAT, Lex: float32(2.71828), Lit: "2.71828"},
		{Pos: tokens.Position{Line: 4, Column: 1}, Tok: tokens.FLOAT, Lex: float32(0), Lit: "0000."},
		{Pos: tokens.Position{Line: 5, Column: 1}, Tok: tokens.FLOAT, Lex: float32(72.4), Lit: "072.40"},
		{Pos: tokens.Position{Line: 6, Column: 1}, Tok: tokens.FLOAT, Lex: float32(0x1.Fp+0), Lit: "0x1.Fp+0"},
		{Pos: tokens.Position{Line: 7, Column: 1}, Tok: tokens.FLOAT, Lex: float32(1.e+0), Lit: "1.e+0"},
		{Pos: tokens.Position{Line: 8, Column: 1}, Tok: tokens.FLOAT, Lex: float32(6.67428e-11), Lit: "6.67428e-11"},
		{Pos: tokens.Position{Line: 9, Column: 1}, Tok: tokens.FLOAT, Lex: float32(1e6), Lit: "1E6"},
		{Pos: tokens.Position{Line: 10, Column: 1}, Tok: tokens.FLOAT, Lex: float32(.25), Lit: ".25"},
		{Pos: tokens.Position{Line: 11, Column: 1}, Tok: tokens.FLOAT, Lex: float32(.12345e+5), Lit: ".12345E+5"},
		{Pos: tokens.Position{Line: 12, Column: 1}, Tok: tokens.FLOAT, Lex: float32(1_5.), Lit: "1_5."},
		{Pos: tokens.Position{Line: 13, Column: 1}, Tok: tokens.FLOAT, Lex: float32(0.15e+0_2), Lit: "0.15e+0_2"},
		{Pos: tokens.Position{Line: 14, Column: 1}, Tok: tokens.FLOAT, Lex: float32(0x1p-2), Lit: "0x1p-2"},
		{Pos: tokens.Position{Line: 15, Column: 1}, Tok: tokens.FLOAT, Lex: float32(0x2.p10), Lit: "0x2.p10"},
		{Pos: tokens.Position{Line: 16, Column: 1}, Tok: tokens.FLOAT, Lex: float32(0x1.Fp+0), Lit: "0x1.Fp+0"},
		{Pos: tokens.Position{Line: 17, Column: 1}, Tok: tokens.FLOAT, Lex: float32(0x.8p-0), Lit: "0X.8p-0"},
		{Pos: tokens.Position{Line: 18, Column: 1}, Tok: tokens.FLOAT, Lex: float32(0x_1FFFp-16), Lit: "0X_1FFFP-16"},
		{Pos: tokens.Position{Line: 19, Column: 1}, Tok: tokens.INT, Lex: int32(350), Lit: "0x15e"},
		{Pos: tokens.Position{Line: 19, Column: 6}, Tok: tokens.SUB, Lex: "-", Lit: "-"},
		{Pos: tokens.Position{Line: 19, Column: 7}, Tok: tokens.INT, Lex: int32(2), Lit: "2"},
	}
	input := readInput("../tests/lexer/test2.txt")
	performTest(t, input, expected[:])
	performTest(t, "1p-2", []tokens.Token{{Pos: tokens.Position{Line: 1, Column: 1}, Tok: tokens.ILLEGAL, Lex: "illegal: p exponent requires hexadecimal mantissa", Lit: ""}})
	performTest(t, "0x1.5e-2", []tokens.Token{{Pos: tokens.Position{Line: 1, Column: 1}, Tok: tokens.ILLEGAL, Lex: "illegal: hexadecimal mantissa requires p exponent", Lit: ""}})
}
func TestIdents(t *testing.T) {
	expected := [...]tokens.Token{
		{Pos: tokens.Position{Line: 1, Column: 1}, Tok: tokens.IDENT, Lex: "test", Lit: "test"},
		{Pos: tokens.Position{Line: 2, Column: 1}, Tok: tokens.INT, Lex: int32(2), Lit: "2"},
		{Pos: tokens.Position{Line: 2, Column: 2}, Tok: tokens.IDENT, Lex: "test", Lit: "test"},
		{Pos: tokens.Position{Line: 3, Column: 1}, Tok: tokens.IDENT, Lex: "test2", Lit: "test2"},
		{Pos: tokens.Position{Line: 4, Column: 1}, Tok: tokens.IDENT, Lex: "Test2", Lit: "Test2"},
		{Pos: tokens.Position{Line: 5, Column: 1}, Tok: tokens.IDENT, Lex: "TEST", Lit: "TEST"},
		{Pos: tokens.Position{Line: 6, Column: 1}, Tok: tokens.IDENT, Lex: "_test2", Lit: "_test2"},
		{Pos: tokens.Position{Line: 7, Column: 1}, Tok: tokens.IDENT, Lex: "_2test", Lit: "_2test"},
		{Pos: tokens.Position{Line: 8, Column: 1}, Tok: tokens.IDENT, Lex: "___", Lit: "___"},
		{Pos: tokens.Position{Line: 9, Column: 2}, Tok: tokens.IDENT, Lex: "αβ", Lit: "αβ"},
	}
	input := readInput("../tests/lexer/test3.txt")
	performTest(t, input, expected[:])
}

func TestComment(t *testing.T) {
	expected := [...]tokens.Token{
		{Pos: tokens.Position{Line: 1, Column: 1}, Tok: tokens.INT, Lex: int32(1), Lit: "1"},
		{Pos: tokens.Position{Line: 1, Column: 3}, Tok: tokens.QUO, Lex: "/", Lit: "/"},
		{Pos: tokens.Position{Line: 1, Column: 5}, Tok: tokens.INT, Lex: int32(2), Lit: "2"},
		{Pos: tokens.Position{Line: 2, Column: 1}, Tok: tokens.COMMENT, Lex: "комментарий 1", Lit: "//комментарий 1"},
		{Pos: tokens.Position{Line: 3, Column: 1}, Tok: tokens.COMMENT, Lex: " 1 + 5 / 10", Lit: "// 1 + 5 / 10"},
		{Pos: tokens.Position{Line: 4, Column: 1}, Tok: tokens.COMMENT, Lex: "comment 1", Lit: "/*comment 1*/"},
		{Pos: tokens.Position{Line: 5, Column: 1}, Tok: tokens.COMMENT, Lex: "/sdfsdf", Lit: "///sdfsdf"},
		{Pos: tokens.Position{Line: 6, Column: 1}, Tok: tokens.COMMENT, Lex: "1t\n2t\n3t", Lit: "/*1t\n2t\n3t*/"},
		{Pos: tokens.Position{Line: 9, Column: 1}, Tok: tokens.IDENT, Lex: "end", Lit: "end"},
		{Pos: tokens.Position{Line: 10, Column: 1}, Tok: tokens.COMMENT, Lex: "comment with no ending quote", Lit: "/*comment with no ending quote"},
	}
	input := readInput("../tests/lexer/test4.txt")
	performTest(t, input, expected[:])
}

func TestOperands(t *testing.T) {
	expected := [...]tokens.Token{
		{Pos: tokens.Position{Line: 1, Column: 1}, Tok: tokens.ADD, Lex: "+", Lit: "+"},
		{Pos: tokens.Position{Line: 1, Column: 6}, Tok: tokens.AND, Lex: "&", Lit: "&"},
		{Pos: tokens.Position{Line: 1, Column: 12}, Tok: tokens.ADD_ASSIGN, Lex: "+=", Lit: "+="},
		{Pos: tokens.Position{Line: 1, Column: 18}, Tok: tokens.AND_ASSIGN, Lex: "&=", Lit: "&="},
		{Pos: tokens.Position{Line: 1, Column: 25}, Tok: tokens.LAND, Lex: "&&", Lit: "&&"},
		{Pos: tokens.Position{Line: 1, Column: 31}, Tok: tokens.EQL, Lex: "==", Lit: "=="},
		{Pos: tokens.Position{Line: 1, Column: 37}, Tok: tokens.NEQ, Lex: "!=", Lit: "!="},
		{Pos: tokens.Position{Line: 1, Column: 43}, Tok: tokens.LPAREN, Lex: "(", Lit: "("},
		{Pos: tokens.Position{Line: 1, Column: 48}, Tok: tokens.RPAREN, Lex: ")", Lit: ")"},
		{Pos: tokens.Position{Line: 2, Column: 1}, Tok: tokens.SUB, Lex: "-", Lit: "-"},
		{Pos: tokens.Position{Line: 2, Column: 6}, Tok: tokens.OR, Lex: "|", Lit: "|"},
		{Pos: tokens.Position{Line: 2, Column: 12}, Tok: tokens.SUB_ASSIGN, Lex: "-=", Lit: "-="},
		{Pos: tokens.Position{Line: 2, Column: 18}, Tok: tokens.OR_ASSIGN, Lex: "|=", Lit: "|="},
		{Pos: tokens.Position{Line: 2, Column: 25}, Tok: tokens.LOR, Lex: "||", Lit: "||"},
		{Pos: tokens.Position{Line: 2, Column: 31}, Tok: tokens.LSS, Lex: "<", Lit: "<"},
		{Pos: tokens.Position{Line: 2, Column: 37}, Tok: tokens.LEQ, Lex: "<=", Lit: "<="},
		{Pos: tokens.Position{Line: 2, Column: 43}, Tok: tokens.LBRACK, Lex: "[", Lit: "["},
		{Pos: tokens.Position{Line: 2, Column: 48}, Tok: tokens.RBRACK, Lex: "]", Lit: "]"},
		{Pos: tokens.Position{Line: 3, Column: 1}, Tok: tokens.MUL, Lex: "*", Lit: "*"},
		{Pos: tokens.Position{Line: 3, Column: 6}, Tok: tokens.XOR, Lex: "^", Lit: "^"},
		{Pos: tokens.Position{Line: 3, Column: 12}, Tok: tokens.MUL_ASSIGN, Lex: "*=", Lit: "*="},
		{Pos: tokens.Position{Line: 3, Column: 18}, Tok: tokens.XOR_ASSIGN, Lex: "^=", Lit: "^="},
		{Pos: tokens.Position{Line: 3, Column: 25}, Tok: tokens.ARROW, Lex: "<-", Lit: "<-"},
		{Pos: tokens.Position{Line: 3, Column: 31}, Tok: tokens.GTR, Lex: ">", Lit: ">"},
		{Pos: tokens.Position{Line: 3, Column: 37}, Tok: tokens.GEQ, Lex: ">=", Lit: ">="},
		{Pos: tokens.Position{Line: 3, Column: 43}, Tok: tokens.LBRACE, Lex: "{", Lit: "{"},
		{Pos: tokens.Position{Line: 3, Column: 48}, Tok: tokens.RBRACE, Lex: "}", Lit: "}"},
		{Pos: tokens.Position{Line: 4, Column: 1}, Tok: tokens.QUO, Lex: "/", Lit: "/"},
		{Pos: tokens.Position{Line: 4, Column: 6}, Tok: tokens.SHL, Lex: "<<", Lit: "<<"},
		{Pos: tokens.Position{Line: 4, Column: 12}, Tok: tokens.QUO_ASSIGN, Lex: "/=", Lit: "/="},
		{Pos: tokens.Position{Line: 4, Column: 18}, Tok: tokens.SHL_ASSIGN, Lex: "<<=", Lit: "<<="},
		{Pos: tokens.Position{Line: 4, Column: 25}, Tok: tokens.INC, Lex: "++", Lit: "++"},
		{Pos: tokens.Position{Line: 4, Column: 31}, Tok: tokens.ASSIGN, Lex: "=", Lit: "="},
		{Pos: tokens.Position{Line: 4, Column: 37}, Tok: tokens.DEFINE, Lex: ":=", Lit: ":="},
		{Pos: tokens.Position{Line: 4, Column: 43}, Tok: tokens.COMMA, Lex: ",", Lit: ","},
		{Pos: tokens.Position{Line: 4, Column: 48}, Tok: tokens.SEMICOLON, Lex: ";", Lit: ";"},
		{Pos: tokens.Position{Line: 5, Column: 1}, Tok: tokens.REM, Lex: "%", Lit: "%"},
		{Pos: tokens.Position{Line: 5, Column: 6}, Tok: tokens.SHR, Lex: ">>", Lit: ">>"},
		{Pos: tokens.Position{Line: 5, Column: 12}, Tok: tokens.REM_ASSIGN, Lex: "%=", Lit: "%="},
		{Pos: tokens.Position{Line: 5, Column: 18}, Tok: tokens.SHR_ASSIGN, Lex: ">>=", Lit: ">>="},
		{Pos: tokens.Position{Line: 5, Column: 25}, Tok: tokens.DEC, Lex: "--", Lit: "--"},
		{Pos: tokens.Position{Line: 5, Column: 31}, Tok: tokens.NOT, Lex: "!", Lit: "!"},
		{Pos: tokens.Position{Line: 5, Column: 37}, Tok: tokens.ELLIPSIS, Lex: "...", Lit: "..."},
		{Pos: tokens.Position{Line: 5, Column: 43}, Tok: tokens.PERIOD, Lex: ".", Lit: "."},
		{Pos: tokens.Position{Line: 5, Column: 48}, Tok: tokens.COLON, Lex: ":", Lit: ":"},
		{Pos: tokens.Position{Line: 6, Column: 6}, Tok: tokens.AND_NOT, Lex: "&^", Lit: "&^"},
		{Pos: tokens.Position{Line: 6, Column: 18}, Tok: tokens.AND_NOT_ASSIGN, Lex: "&^=", Lit: "&^="},
	}
	input := readInput("../tests/lexer/test5.txt")
	performTest(t, input, expected[:])
}

func TestChar(t *testing.T) {
	expected := [...]tokens.Token{
		{Pos: tokens.Position{Line: 1, Column: 1}, Tok: tokens.CHAR, Lex: "a", Lit: "'a'"},
		{Pos: tokens.Position{Line: 2, Column: 1}, Tok: tokens.CHAR, Lex: "ä", Lit: "'ä'"},
		{Pos: tokens.Position{Line: 3, Column: 1}, Tok: tokens.CHAR, Lex: "本", Lit: "'本'"},
		{Pos: tokens.Position{Line: 4, Column: 1}, Tok: tokens.CHAR, Lex: "\t", Lit: "'\t'"},
		{Pos: tokens.Position{Line: 5, Column: 1}, Tok: tokens.CHAR, Lex: "\u12e4", Lit: "'\u12e4'"},
		{Pos: tokens.Position{Line: 6, Column: 1}, Tok: tokens.CHAR, Lex: "\U00101234", Lit: "'\U00101234'"},
	}
	input := readInput("../tests/lexer/test6.txt")
	performTest(t, input, expected[:])
	performTest(t, "'aa'", []tokens.Token{{Pos: tokens.Position{Line: 1, Column: 1}, Tok: tokens.ILLEGAL, Lex: "illegal: rune literal not terminated", Lit: ""}})
}

func TestString(t *testing.T) {
	expected := [...]tokens.Token{
		{Pos: tokens.Position{Line: 1, Column: 1}, Tok: tokens.STRING, Lex: "abc", Lit: "`abc`"},
		{Pos: tokens.Position{Line: 2, Column: 1}, Tok: tokens.STRING, Lex: "\n\n\n", Lit: "`\n\n\n`"},
		{Pos: tokens.Position{Line: 6, Column: 1}, Tok: tokens.STRING, Lex: "\n", Lit: "\"\\n\""},
		{Pos: tokens.Position{Line: 7, Column: 1}, Tok: tokens.STRING, Lex: `"`, Lit: `"\""`},
		{Pos: tokens.Position{Line: 8, Column: 1}, Tok: tokens.STRING, Lex: "Hello, world!\n", Lit: `"Hello, world!\n"`},
		{Pos: tokens.Position{Line: 9, Column: 1}, Tok: tokens.STRING, Lex: `日本語`, Lit: `"日本語"`},
	}
	input := readInput("../tests/lexer/test9.txt")
	performTest(t, input, expected[:])
	expected2 := [...]tokens.Token{
		{Pos: tokens.Position{Line: 1, Column: 1}, Tok: tokens.STRING, Lex: "日本語", Lit: `"\u65e5本\U00008a9e"`},
		{Pos: tokens.Position{Line: 2, Column: 2}, Tok: tokens.STRING, Lex: "ÿÿ", Lit: `"\xff\u00FF"`},
	}
	const input2 = `"\u65e5本\U00008a9e"
	"\xff\u00FF"`
	performTest(t, input2, expected2[:])
	performTest(t, `"\uD800"`, []tokens.Token{{Pos: tokens.Position{Line: 1, Column: 1}, Tok: tokens.ILLEGAL, Lex: "illegal: invalid Unicode code point", Lit: ""}})
	performTest(t, `"\U00110000"`, []tokens.Token{{Pos: tokens.Position{Line: 1, Column: 1}, Tok: tokens.ILLEGAL, Lex: "illegal: invalid Unicode code point", Lit: ""}})
}

func TestStringFormats(t *testing.T) {
	expected := [...]tokens.Token{
		{Pos: tokens.Position{Line: 1, Column: 1}, Tok: tokens.STRING, Lex: "日本語", Lit: `"日本語"`},
		{Pos: tokens.Position{Line: 1, Column: 7}, Tok: tokens.STRING, Lex: "日本語", Lit: `"\u65e5\u672c\u8a9e"`},
		{Pos: tokens.Position{Line: 1, Column: 28}, Tok: tokens.STRING, Lex: "日本語", Lit: `"\U000065e5\U0000672c\U00008a9e"`},
	}
	input := readInput("../tests/lexer/test7.txt")
	performTest(t, input, expected[:])
}

func TestHelloWorld(t *testing.T) {
	expected := []tokens.Token{
		{Pos: tokens.Position{Line: 1, Column: 1}, Tok: tokens.PACKAGE, Lex: "package", Lit: "package"},
		{Pos: tokens.Position{Line: 1, Column: 9}, Tok: tokens.IDENT, Lex: "hello", Lit: "hello"},
		{Pos: tokens.Position{Line: 3, Column: 1}, Tok: tokens.IMPORT, Lex: "import", Lit: "import"},
		{Pos: tokens.Position{Line: 3, Column: 8}, Tok: tokens.LPAREN, Lex: "(", Lit: "("},
		{Pos: tokens.Position{Line: 4, Column: 5}, Tok: tokens.STRING, Lex: "fmt", Lit: `"fmt"`},
		{Pos: tokens.Position{Line: 5, Column: 1}, Tok: tokens.RPAREN, Lex: ")", Lit: ")"},
		{Pos: tokens.Position{Line: 6, Column: 1}, Tok: tokens.COMMENT, Lex: "\nsimple programm that greets you!\naccepts nothing\n", Lit: "/*\nsimple programm that greets you!\naccepts nothing\n*/"},
		{Pos: tokens.Position{Line: 10, Column: 1}, Tok: tokens.FUNC, Lex: "func", Lit: "func"},
		{Pos: tokens.Position{Line: 10, Column: 6}, Tok: tokens.IDENT, Lex: "main", Lit: "main"},
		{Pos: tokens.Position{Line: 10, Column: 10}, Tok: tokens.LPAREN, Lex: "(", Lit: "("},
		{Pos: tokens.Position{Line: 10, Column: 11}, Tok: tokens.RPAREN, Lex: ")", Lit: ")"},
		{Pos: tokens.Position{Line: 10, Column: 13}, Tok: tokens.LBRACE, Lex: "{", Lit: "{"},
		{Pos: tokens.Position{Line: 11, Column: 5}, Tok: tokens.CONST, Lex: "const", Lit: "const"},
		{Pos: tokens.Position{Line: 11, Column: 11}, Tok: tokens.IDENT, Lex: "message", Lit: "message"},
		{Pos: tokens.Position{Line: 11, Column: 19}, Tok: tokens.ASSIGN, Lex: "=", Lit: "="},
		{Pos: tokens.Position{Line: 11, Column: 21}, Tok: tokens.STRING, Lex: "Hello world!\nend of the message", Lit: "`Hello world!\nend of the message`"},
		{Pos: tokens.Position{Line: 13, Column: 5}, Tok: tokens.IDENT, Lex: "fmt", Lit: "fmt"},
		{Pos: tokens.Position{Line: 13, Column: 8}, Tok: tokens.PERIOD, Lex: ".", Lit: "."},
		{Pos: tokens.Position{Line: 13, Column: 9}, Tok: tokens.IDENT, Lex: "Printf", Lit: "Printf"},
		{Pos: tokens.Position{Line: 13, Column: 15}, Tok: tokens.LPAREN, Lex: "(", Lit: "("},
		{Pos: tokens.Position{Line: 13, Column: 16}, Tok: tokens.IDENT, Lex: "message", Lit: "message"},
		{Pos: tokens.Position{Line: 13, Column: 23}, Tok: tokens.RPAREN, Lex: ")", Lit: ")"},
		{Pos: tokens.Position{Line: 14, Column: 1}, Tok: tokens.RBRACE, Lex: "}", Lit: "}"},
	}

	input := readInput("../tests/lexer/test8.txt")
	performTest(t, input, expected[:])
}
