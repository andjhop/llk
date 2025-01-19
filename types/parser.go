package types

import (
	"strconv"
	"text/scanner"
)

// Parser is the basic interface for a parser. A Parser is simply an
// object which defines a Parse method accepting a lexical stream
// emitted by a Tokensier and ultimately produces some result.
type Parser[V any] interface {
	// Name returns the name of the parser, used in
	// error messages
	Name() string

	// Parse "executes" the parser, applying the parser
	// to the token stream input emmitted by a
	// Tokeniser. Parse Returns a Result respresenting
	// the reulting outcome: a fail, with a Result
	// containing a set of parse errors; or a success,
	// with a Result containing a  user defined value
	// and set of locations where the parser
	// successfully finished recognising a sequence of
	// tokens
	Parse(Tokeniser) Result[V]
}

// ScanErrType
type ScanErrType int

const (
	// ScanErrEOF
	ScanErrEOF ScanErrType = -(1 + iota)

	// ScanErrMsg
	ScanErrMsg
)

// ScanErr
type ScanErr struct {
	ErrType ScanErrType
	ErrMsg  string
}

func NewScanErr(errType ScanErrType, errMsg string) ScanErr {
	return ScanErr{
		errType,
		errMsg,
	}
}

// Tokeniser represents a lexical scanner or tokensier which "emits"
// Tokens and whose whose current location returned by Loc(), called k
// is the kth Token to be scanned. Tokensier can be moved to an abitrary
// location >= 0 and less than the higest index k of the most recent
// token to be scanned
type Tokeniser interface {
	Line() int
	Column() int

	// Loc returns the current location of the
	// Tokeniser
	Loc() (k int)

	// Dec moves the Tokensier to the previous location
	// in the token stream, calling Dec to move before
	// the "begning" of the token stream is an error and
	// should result in a panic
	Dec()

	// Inc moves the Tokensier to the next location in
	// the token stream, calling Inc to move beyond the
	// "end" of the token stream is an error and should
	// result in a panic
	Inc()

	// Seek moves the Tokeniser to an arbitrary location,
	// or the kth location in the token stream
	Seek(k int)

	// Peek returns the the Token at the current
	// location of the tokeniser without actually
	// advancing the location
	Peek() (Token, ScanErr)
}

// Token represents a lexical token emitted by a Tokeniser. A tokeniser
// has an associated lexical category which defines its "class" or
// "meaning"; the class of its match string
type Token struct {
	// category is the lexical category this token
	// belongs to
	category rune

	// match is the actual token value matched from the
	// tokeniser input text
	match string
}

func NewToken(c rune, m string) Token {
	return Token{
		category: c,
		match:    m,
	}
}

// Empty is is the most primitive parser. It only recognises the Empty
// string and so always succeeds returning a result containing the
// scanner location and the user defined value v
type Empty[V any] struct {
	value V
}

func NewEmpty[V any](v V) Empty[V] {
	return Empty{v}
}

func (Empty[V]) Name() string {
	return ""
}

// Parse represents a lexical token emitted by a Tokeniser. A tokeniser
// has an associated lexical category which defines its "class" or
func (e Empty[V]) Parse(s Tokeniser) Result[V] {
	return NewSucceeded(e.value, s.Loc())
}

// converter or converters are, functions called to convert the token
// text. A converter take the token text as input and returns and any
// and possibly and error indicating that the conversion failed
type converter func(string) (any, error)

// Term is the most primitive parser that can fail or succeed. A Term
// parser recognises the next token and returns a true parse result if
// the lexeme matches the lexical category specified by category. value
// can be optionally specified to require a specific match for the
// parsed token value.  name is the name of the Parser used in parse
// error messages. Calling Parse.
type Term[V any] struct {
	name string

	// category is the lexical category that this
	// parser recognises
	category rune

	// exactMatch is a list a list of strings this
	// parser is allowed to recognise beyond only
	// the lexical category
	exactMatch string

	// converter is called to convert the literal
	// token text matched by this parser into the actual
	// value stored in the Term's parse result
	converter converter
}

func NewTerm[V any](name string, category rune) Term[V] {
	return Term{
		name:       name,
		category:   category,
		exactMatch: "",
		converter: func(s string) (any, error) {
			return s, nil
		},
	}
}

// Text returns a Parser which parses a unicode character and only
// succeeds if the parsed token text matches the character specified by
// the category
func Text(category rune) Term[string] {
	return NewTerm(string(category), category)
}

// Id returns a Parser which parsers a go idenitfier and only succeeds
// if the parsed token text exactly matches the string specified by s
func Id(s string) Term[string] {
	return NewTerm("identifer", scanner.Ident).
		WithExactMatch(s)
}

// Int returns a Parser which parsers a go decimal literal and returns
// an returns the corresponding value as an int64 in the parser
// result
func Int() Term[int] {
	return NewTerm("integer", scanner.Int).
		WithConverter(func(s string) (any, error) {
			return strconv.ParseInt(s, 10, 64)
		})
}

// Int returns a Parser which parsers a go floating point literal and
// returns an returns the corresponding value as an float64 in the
// parser result
func Float() Term[float64] {
	return NewTerm("float", scanner.Float).
		WithConverter(func(s string) (any, error) {
			return strconv.ParseFloat(s, 64)
		})
}

// String returns a Parser which parsers a go quoted string literal and
// returns an returns the corresponding value as astring
func String() Term[string] {
	return NewTerm("quoted string", scanner.String).
		WithConverter(func(s string) (any, error) {
			return strconv.Unquote(s)
		})
}

// WithExactmatch returns a Term which has to match the exact token
// text specified by s and will otherwise fail
func (t Term[V]) WithExactMatch(s string) Term[V] {
	t.exactMatch = s
	return t
}

// WithConverter returns a Term which calls the converter c on the token
// text and stores the returend value instead of the token text itself
func (t Term[V]) WithConverter(c converter) Term[V] {
	t.converter = c
	return t
}

func (t Term[V]) Name() string {
	return t.name
}

// Parse takes a Text and returns a LocSet representing the NewTerm
// returns a Terminal with the name n, which matches a token of the
// lexical category specified by c.
func (t Term[V]) Parse(tokeniser Tokeniser) Result[V] {
	tokLine, tokCol := tokeniser.Line(), tokeniser.Column()

	token, scanErr := tokeniser.Peek()
	switch scanErr.ErrType {
	case ScanErrEOF:
		return NewHalt("scanner", "end of file").
			WithLineAndColumn(tokLine, tokCol)
	case ScanErrMsg:
		return NewHalt("scanner", scanErr.ErrMsg).
			WithLineAndColumn(tokLine, tokCol)
	}
	switch {
	case token.category != t.category:
		fallthrough
	case t.exactMatch != "" && token.match != t.exactMatch:
		return NewScanFailed(t.name, string(token.match),
			tokLine, tokCol)
	}
	v, err := t.converter(token.match)
	if err != nil {
		return NewHalt("conversion", "went wrong").
			WithLineAndColumn(tokLine, tokCol)
	}
	tokeniser.Inc()
	return NewSucceeded(v, tokeniser.Loc())
}
