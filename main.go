package main

import (
	"fmt"
	lexer "gocompiler/src"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("input.txt")
	if err != nil {
		panic(err)
	}
	lexerInstance := lexer.NewLexer(file)
	for {
		pos, tok, lex, lit := lexerInstance.Lex()
		if tok == lexer.EOF {
			break
		}
		fmt.Printf("%d:%d\t%s\t%s\t%s\n", pos.Line, pos.Column, tok, strings.ReplaceAll(lex, "\r", ""), strings.ReplaceAll(lit, "\r", ""))
	}
}
