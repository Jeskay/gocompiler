package parser

import (
	"fmt"
	"gocompiler/src/lexer"
	"gocompiler/src/tokens"
	"io"
	"os"
	"strings"
	"testing"
)

func readInput(filename string) string {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0600)
	if err != nil {
		panic(err)
	}
	b, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func performTest(t *testing.T, input string, expect string) {
	lexerInstance := lexer.NewLexer(strings.NewReader(input))
	var tokenList []tokens.Token
	for {
		pos, tok, lex, lit := lexerInstance.Lex()
		tokenList = append(tokenList, tokens.Token{Pos: pos, Tok: tok, Lex: lex, Lit: lit})
		if tok == tokens.EOF || tok == tokens.ILLEGAL {
			break
		}
	}
	parserInstance := NewParser(tokenList)
	astTree := parserInstance.Parse()
	result := PrintAST(astTree)
	if result != expect {
		t.Errorf("expected %s got %s", expect, result)
	}
}

func testPath(path string, input bool) string {
	if input {
		return "../tests/parser/input/" + path + "/test"
	} else {
		return "../tests/parser/output/" + path + "/test"
	}
}
func runTestFolder(t *testing.T, path string, amount int) {
	for i := 1; i <= amount; i++ {
		input := readInput(testPath(path, true) + fmt.Sprint(i) + ".txt")
		expected := readInput(testPath(path, false) + fmt.Sprint(i) + ".txt")
		performTest(t, input, expected)
	}
}

func TestFunctions(t *testing.T) {
	const testAmount = 4
	const path = "functions"
	runTestFolder(t, path, testAmount)
}

func TestVarDeclarations(t *testing.T) {
	runTestFolder(t, "var", 3)
}

func TestStructs(t *testing.T) {
	const testAmount = 3
	const path = "structs"
	runTestFolder(t, path, testAmount)
}
