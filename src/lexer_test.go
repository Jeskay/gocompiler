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
func TestDigits(t *testing.T) {
	expected := [...]lexem{
		{Position{1, 1}, INT, "1910", "1910"},
		{Position{2, 1}, INT, "0", "0"},
		{Position{3, 1}, INT, "4", "0b100"},
		{Position{4, 1}, INT, "7", "0b00111"},
		{Position{5, 1}, INT, "511", "0777"},
		{Position{6, 1}, INT, "668", "0o1234"},
		{Position{7, 1}, INT, "282", "0O0432"},
		{Position{8, 1}, INT, "427", "0x01AB"},
		{Position{9, 1}, INT, "171", "0Xab"},
	}
	const input = "1910 \n0 \n0b100 \n0b00111 \n0777 \n0o1234 \n0O0432 \n0x01AB \n0Xab"
	lexerInstance := NewLexer(strings.NewReader(input))
	for i := 0; ; i++ {
		pos, tok, lex, lit := lexerInstance.Lex()
		if tok == EOF {
			break
		}
		got := lexem{pos, tok, lex, lit}
		if !got.Compare(expected[i]) {
			t.Errorf("expected %s, got %s", expected[i].ToString(), got.ToString())
		}
	}

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
	lexerInstance := NewLexer(strings.NewReader(input))
	for i := 0; ; i++ {
		pos, tok, lex, lit := lexerInstance.Lex()
		if tok == EOF {
			break
		}
		got := lexem{pos, tok, lex, lit}
		if !got.Compare(expected[i]) {
			t.Errorf("expected %s, got %s", expected[i].ToString(), got.ToString())
		}
	}
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
	lexerInstance := NewLexer(strings.NewReader(input))
	for i := 0; ; i++ {
		pos, tok, lex, lit := lexerInstance.Lex()
		if tok == EOF {
			break
		}
		got := lexem{pos, tok, lex, lit}
		if !got.Compare(expected[i]) {
			t.Errorf("expected %s, got %s", expected[i].ToString(), got.ToString())
		}
	}
}
