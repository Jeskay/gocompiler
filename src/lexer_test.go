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
