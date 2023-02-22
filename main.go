package main

import (
	"encoding/json"
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
		lexerInstance := lexer.NewLexer(file)
		var tokenList []tokens.Token
		for {
			pos, tok, lex, lit := lexerInstance.Lex()
			if tok == tokens.EOF {
				break
			}
			tokenList = append(tokenList, tokens.Token{Pos: pos, Tok: tok, Lex: lex, Lit: lit})
			if tok == tokens.ILLEGAL {
				break
			}
		}
		parserInstance := parser.NewParser(tokenList)
		astTree := parserInstance.Parse()
		for _, node := range astTree {
			s, _ := json.MarshalIndent(node, "", "\t")
			fmt.Println(s)
		}
	}
}
