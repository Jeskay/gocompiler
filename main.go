package main

import (
	"flag"
	"fmt"
	lexer "gocompiler/src/lexer"
	"gocompiler/src/parser"
	"gocompiler/src/tokens"
	"os"
	"strings"
)

var options struct {
	lex    bool
	ast    bool
	source string
}

func main() {
	//var t, err = strconv.ParseFloat("0x2.p10", 32)
	//fmt.Print(t, err)
	flag.StringVar(&options.source, "source", "input.txt", "filename of source to lex")
	flag.BoolVar(&options.lex, "lex", false, "perform lexical analysis")
	flag.BoolVar(&options.ast, "ast", false, "creates AST tree for the code")
	flag.Parse()

	if options.lex {
		file, err := os.Open(options.source)
		if err != nil {
			panic(err)
		}
		lexerInstance := lexer.NewLexer(file)
		for {
			pos, tok, lex, lit := lexerInstance.Lex()
			if tok == tokens.EOF {
				break
			}
			fmt.Printf("%d:%d\t%s\t%v\t%s\n", pos.Line, pos.Column, tok, lex, strings.ReplaceAll(lit, "\r", ""))
			if tok == tokens.ILLEGAL {
				break
			}
		}
	} else if options.ast {
		file, err := os.Open(options.source)
		if err != nil {
			panic(err)
		}
		parserInstance := parser.NewParser(file)
		astTree := parserInstance.Parse()
		result := parser.ResolveFile(astTree, func(pos tokens.Position, msg string) {
			fmt.Println("declaration error: " + msg)
		})
		str := parser.PrintAST(astTree)
		fmt.Println(str)
		fmt.Println(result)
	}
}
