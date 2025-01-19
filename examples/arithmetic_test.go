package examples

import (
	"fmt"
	"strings"
	"testing"

	k "llk"
	"llk/types"
)

func formatFailed(result types.Result) (out string) {
	switch result.(type) {
	case types.Halt:
		halt := result.(types.Halt)
		out += fmt.Sprintf("%d:%d ", halt.Line, halt.Column)
		out += halt.Message
	case types.Failed:
		failed := result.(types.Failed)
		parseErrors := failed.Errors()
		for _, parseErr := range parseErrors {
			if parseErr.Expected == "" || parseErr.Found == "" {
				continue
			}
			out += fmt.Sprintf("\n-\t * %d:%d: wanted: %s, got: `%s`",
				parseErr.Line, parseErr.Column, parseErr.Expected, parseErr.Found)
		}
		out = "Buggy expression: parse error:" + out
	}
	return
}

func TestArithmetic(t *testing.T) {
	// arithmetic expression parser for the grammar:
	//
	//	<expr> → <int> | <subexpr>
	//	<subexpr> → `(` <expr> `+` <expr> `)`
	var (
		expr k.Chain
	)
	tokeniser := k.NewTokeniser(strings.NewReader(
		//"(1 + 2)",
		`((1 +
		(2 + (3 + 4))) +
		(2 + (3 + 4)))`,
	))

	// sub expression parser
	//
	//	<subexpr> → `(` <expr> `+` <expr> `)`
	subexpr := k.
		SeqText("subexpr", '(').
		Lazy(func(any) k.Parser {
			return expr
		}).
		Text('+').
		Lazy(func(a any) k.Parser {
			return k.Seq("subexpr", expr).
				Return(func(b any) any {
					return a.(int64) + b.(int64)
				})
		}).
		Text(')')

	// expr parser
	//
	//	<expr> → <int> | <subexpr>
	expr = k.
		EitherInt("expr").
		Chain(subexpr)

	//t.Error(expr.Parse(tokeniser))
	result := expr.Parse(tokeniser)
	//_, msg, ok := tokeniser.Peek()
	//if msg != "reached eof" {
	//t.Error("error unconsumed input", ok)
	//}
	//t.Error("tokenizer err:", ok, msg)
	t.Error("---", formatFailed(result))
	t.Error(result)
}
