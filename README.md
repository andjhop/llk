# llk

[![Go Reference](https://pkg.go.dev/badge/github.com/andjam/btree.svg)](https://pkg.go.dev/github.com/andjhop/llk)

llk is a [parser combinator](https://en.wikipedia.org/wiki/Parser_combinator) library for building
LL(\*) parsers. Without the jargon, this means that llk offers a bunch of primitives and ways to
combine primitives to construct complex parsers from less complex parsers in such a way the the
result is always  LL(\*). The following is an example of how an recogniser for arithmetic
expressions might be constructed:
```go
p = k.
    EitherInt().
    Chain(k.
        SeqText('(').
            Lazy(func(any) k.Parser {
                return p
            }).
            Text('+').
            Lazy(func(any) k.Parser {
                return p
            }).
            Text(')'))
```


### Scanning and Lexical Rules and Types
The implied lexical structure of any text input into an llk parser is just as you'd expect from a Go
program. llk uses the standard library package `scanner.Scanner` internally and so recognises the
same lexical elements as defined in the go language spec, skipping all whitespace and comments:

* `scanner.Ident` An Identier; just a sequence of one more more letters and digits
* `scanner.Int` Integer literals representing integer constants  
* `scanner.Float` Floating-point literals representing floating-point constants  
* `scanner.String` String literals represents character sequences

### Parsing and Parser Combinator Types
Primitive parsers parse a single lexical element and either succeed or fail. The two primitive most
primitive parsers `Empty` and `Term` can be created using:

* `NewEmpty(v any)` Returns a new parser that recognises the empty string and so always succeeds
  with the result value v.
  
* `NewTerm(category rune)` Returns a new parser that recognises a token of the given category
  and only succeeds if there is a match

To ease the creation of primitives, the following shorthands for creating terminal parsers are
available:
* `Text()` Returns a terminal parser which parses a unicode character
  
* `Id(string)` Returns a terminal parser which recognises a single identifier
  
* `Int()` Returns a terminal parser which recognises a single integer literal and returns th
  corresponding value as an int64
  
* `Float()` Returns a terminal parser which parsers a go floating point literal and returns an
  returns the corresponding value as a float64
  
* `String()` Returns a terminal parser  which parsers a  quoted string literal

These primitives can be combined using the combinators or "chainable" constructors: `Seq` and `Either`. This is a parser
defined by the type: `Chain` and the corresponding operations: `a.Chain(b)` which returns a new
Parser composed of two parsers `a` and `b`;  and `a.Lazy(func(any) Parser { return b })` which does the
same but allows the choice of be to be deferred until execution time. The two primary "chainables"
constructors are:

* `Seq(Parser)` Which combines subsequent chained parsers, returning a new Parser which requires all
  constituent parsers to succeed in sequence. For example, the primitive parsers `Id('a')` which
  recognises the input text `a`, and  `Id('b')` which recognise the input text `b` can be combined
  using `Seq(Id('a')).Chain(Id('b'))` to construct a new parser which regonises the input `ab`.
  
* `Either(Parser)` Which combines constituent parsers, returning an new parser which succeeds if
  even a single of its constituent parsers succeeds. `Either(Id('a')).Chain(Id('b'))` returns a
  parser which succeeds on either of the inputs `a` or `b`.
