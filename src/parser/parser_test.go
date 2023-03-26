package parser

import (
	"fmt"
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
	return strings.ReplaceAll(string(b), "\r", "")
}

func performTest(t *testing.T, input string, expect string) {
	parserInstance := NewParser(strings.NewReader(input))
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

func TestArrays(t *testing.T) {
	runTestFolder(t, "arrays", 2)
}

func TestIfStatements(t *testing.T) {
	runTestFolder(t, "if_statements", 2)
}

func TestForStatements(t *testing.T) {
	runTestFolder(t, "for_statements", 5)
}

func TestEpxressions(t *testing.T) {
	runTestFolder(t, "expressions", 4)
}
