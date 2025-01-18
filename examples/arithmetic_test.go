package examples

import (
	"strings"
	"testing"

	"llk"
)

func TestArithmetic(t *testing.T) {
	// arithmetic expression parser for the grammar:
	//
	//	<expr> → <int> | <subexpr>
	//	<subexpr> → `(` <expr> `+` <expr> `)`
	var (
		expr llk.Chain
	)
	tokeniser := llk.NewTokeniser(strings.NewReader(
		"((1 + (2 + (3 + 4))) + (2 + (3 + 4)))",
	))

	// sub expression parser
	//
	//	<subexpr> → `(` <expr> `+` <expr> `)`
	subexpr := llk.
		SeqText("subexpr", '(').
		Lazy(func(any) llk.Parser {
			return expr
		}).
		Text('+').
		Lazy(func(a any) llk.Parser {
			return llk.Seq("", expr).
				Return(func(b any) any {
					return a.(int64) + b.(int64)
				})
		}).
		Text(')')

	// expr parser
	//
	//	<expr> → <int> | <subexpr>
	expr = llk.
		EitherInt("expr").
		Chain(subexpr)

	t.Error(expr.Parse(tokeniser))
}
