package main

import (
	"flag"
	"fmt"
	lexer "gocompiler/src"
	"os"

	//"strconv"
	"strings"
)

var options struct {
	lex    bool
	source string
}

func main() {
	//var t, err = strconv.ParseFloat("0x2.p10", 32)
	//fmt.Print(t, err)
	flag.StringVar(&options.source, "source", "input.txt", "filename of source to lex")
	flag.BoolVar(&options.lex, "lex", false, "perform lexical analysis")
	flag.Parse()

	if options.lex {
		file, err := os.Open(options.source)
		if err != nil {
			panic(err)
		}
		lexerInstance := lexer.NewLexer(file)
		for {
			pos, tok, lex, lit := lexerInstance.Lex()
			if tok == lexer.EOF {
				break
			}
			fmt.Printf("%d:%d\t%s\t%v\t%s\n", pos.Line, pos.Column, tok, lex, strings.ReplaceAll(lit, "\r", ""))
			if tok == lexer.ILLEGAL {
				break
			}
		}
	}
}
