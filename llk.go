// package llk is a parser combinator library for building LL(finite)
// parsers. Without the jargon, this means that llk offers a bunch of
// primitives and ways to combine primitives to construct complex
// parsers from more primitive ones in such a way the result is
// always LL(finite)
package llk

import (
	"strconv"
	"strings"
	"text/scanner"

	"llk/types"
)

// Parser is the basic interface for a parser. A Parser is simply an
// object which defines a Parse method accepting a lexical stream
// emitted by a Tokensier and ultimately produces some result.
type Parser = types.Parser

// Chain is a Parser which is a composition or combination of other
// other Primitive or Chained parsers. That is, Chain implements the
// Parser interface but offers extra functionality for combining parsers
// together. The exact meaning or way in which parsers are combined is
// determined by a "folder"
type Chain = *types.M

// tokeniser implements of the Tokeniser interface, emmiting tokens from
// the scanner and storing scanned tokens in tokens
type tokeniser struct {
	scanner *scanner.Scanner

	// tokens is the sequence of tokens already scanned
	// by the tokeniser, the kth token to be scanned is
	// stored at the kth index of tokens
	tokens []types.Token

	// loc is the "current" location of the tokeniser,
	// returned by Loc()
	loc int
}

func newTokeniser(r *strings.Reader) tokeniser {
	s := &scanner.Scanner{}
	s.Init(r)

	return tokeniser{
		scanner: s,
	}
}

// Loc returns the current location of the Tokeniser
func (t tokeniser) Loc() int {
	return t.loc
}

// Dec moves the Tokensier to the previous location in the token stream,
// calling Dec to move before the "begning" of the token stream is an
// error and results in a panic
func (t *tokeniser) Dec() {
	t.loc--
}

// Inc moves the Tokensier to the next location in the token stream,
// calling Inc to move beyond the "end" of the token stream is an error
// and results in a panic
func (t *tokeniser) Inc() {
	t.loc++
}

// Seek moves the location of the scanner to some arbitrary point in the
// location of the scanner to some arbitrary point in the past
func (t *tokeniser) Seek(loc int) {
	if loc < 0 || loc > len(t.tokens) {
		panic(types.ErrBadLoc)
	}
	t.loc = loc
}

// Peek returns the Token at the current location of the tokeniser
// without actually advancing the location. Peak also returns the flag
// ok, indicating whether or not we reached the end of the input
func (t *tokeniser) Peek() (token types.Token, ok bool) {
	if t.loc >= len(t.tokens) {
		category := t.scanner.Scan()
		if category == scanner.EOF {
			return
		}
		t.tokens = append(
			t.tokens,
			types.NewToken(category, t.scanner.TokenText()),
		)
	}
	return t.tokens[t.loc], true
}

// Text returns a Parser which parses a unicode character and only
// succeeds if the parsed token text matches the character specified by
// the category
func Text(category rune) types.Term {
	return types.NewTerm("text", category)
}

// Id returns a Parser which parsers a go idenitfier and only succeeds
// if the parsed token text exactly matches the string specified by s
func Id(s string) types.Term {
	return types.
		NewTerm("identifer", scanner.Ident).
		WithExactMatch(s)
}

// Int returns a Parser which parsers a go decimal literal and returns
// an returns the corresponding value as an int64 in the parser
// result
func Int() types.Term {
	return types.
		NewTerm("integer", scanner.Int).
		WithConverter(func(s string) (any, error) {
			return strconv.ParseInt(s, 10, 64)
		})
}

// Int returns a Parser which parsers a go floating point literal and
// returns an returns the corresponding value as an float64 in the
// parser result
func Float() types.Term {
	return types.
		NewTerm("float", scanner.Float).
		WithConverter(func(s string) (any, error) {
			return strconv.ParseFloat(s, 64)
		})
}

// String returns a Parser which parsers a go quoted string literal and
// returns an returns the corresponding value as astring
func String() types.Term {
	return types.
		NewTerm("quoted string", scanner.String).
		WithConverter(func(s string) (any, error) {
			return strconv.Unquote(s)
		})
}

// Seq returns a chainable parser which applies parsers in sequence to
// the input token stream. That is, it applies the first parser a to the
// input, and for each finishing location, applies the next parser,
// parser b. The following:
//
//	Seq(Id('a')).
//		Chain(Id('b')).
//		Chain(Id('c')).
//
// Is a parser which combines the primitve parsers Id('a'), Id('b') and
// Id('c') to construct a parser which parsers the text "abc". Parsers
// can be defined recursively using Lazy instead of chain:
//
//	 var p Chainable
//
//		p = Seq(Id('a')).Lazy(func (any) Parser {
//			return p.Lazy(func(any) Parser {
//				return Id('b')
//			})
//		})
//
// Is a parser which parsers the input text: a*b*. This can be unnested,
// and so define equivalently like:
//
//	 var p Chainable
//
//	p = Seq(Id('a')).Lazy(func (any) Parser {
//		return p
//	}).Lazy(func(any) Parser {
//		return Id('b')
//	})
func Seq(n string, p types.Parser) Chain {
	return types.NewM(func(c Chain, s types.Tokeniser) (r types.Result) {
		switch c.Result().(type) {
		case types.Failed:
			r = c.Result()
		case types.Succeeded:
			r = types.NewFailed("")
		}
		for loc := range c.Result().Locs() {
			s.Seek(loc)
			r = r.Join(c.Parse(s))
		}
		return
	}).WithName(n).Chain(p)
}

// Either returns a chainable parser which applies all parsers to the
// input text and succeeds if any one parser succeedes, the following:
//
//	Either(Id('a')).
//		Chain(Id('b')).
//		Chain(Id('c')).
//
// Is a parser which parsers any of the inputs "a", "b", or "c"
func Either(n string, p types.Parser) Chain {
	return types.NewM(func(c Chain, s types.Tokeniser) (r types.Result) {
		r = c.Result().Join(c.Parse(s))
		return
	}).WithName(n).Chain(p)
}
