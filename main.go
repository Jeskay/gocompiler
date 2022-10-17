package main

import (
	"fmt"
	lexer "gocompiler/src"
	"os"
)

func main() {
	file, err := os.Open("input.txt")
	if err != nil {
		panic(err)
	}
	lexerInstance := lexer.NewLexer(file)
	for {
		pos, tok, lit := lexerInstance.Lex()
		if tok == lexer.EOF {
			break
		}

		fmt.Printf("%d:%d\t%s\t%s\n", pos.Line, pos.Column, tok, lit)
	}
}