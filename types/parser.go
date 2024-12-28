package types

// Parser is the basic interface for a parser. A Parser is simply an
// object which defines a Parse method accepting a lexical stream
// emitted by a Tokensier and ultimately produces some result.
type Parser interface {
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
	Parse(Tokeniser) Result
}

// Tokeniser represents a lexical scanner or tokensier which "emits"
// Tokens and whose whose current location returned by Loc(), called k
// is the kth Token to be scanned. Tokensier can be moved to an abitrary
// location >= 0 and less than the higest index k of the most recent
// token to be scanned
type Tokeniser interface {
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
	Peek() (Token, bool)
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

func NewToken(t rune, match string) Token {
	return Token{t, match}
}

// Empty is is the most primitive parser. It only recognises the Empty
// string and so always succeeds returning a result containing the
// scanner location and the user defined value v
type Empty struct {
	value any
}

func NewEmpty(v any) Empty {
	return Empty{v}
}

func (Empty) Name() string {
	return ""
}

// Parse represents a lexical token emitted by a Tokeniser. A tokeniser
// has an associated lexical category which defines its "class" or
func (e Empty) Parse(s Tokeniser) Result {
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
type Term struct {
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

func NewTerm(name string, category rune) Term {
	return Term{
		name:       name,
		category:   category,
		exactMatch: "",
		converter: func(s string) (any, error) {
			return s, nil
		},
	}
}

// WithExactmatch returns a Term which has to match the exact token
// text specified by s and will otherwise fail
func (t Term) WithExactMatch(s string) Term {
	t.exactMatch = s
	return t
}

// WithConverter returns a Term which calls the converter c on the token
// text and stores the returend value instead of the token text itself
func (t Term) WithConverter(c converter) Term {
	t.converter = c
	return t
}

func (t Term) Name() string {
	return t.name
}

// Parse takes a Text and returns a LocSet representing the NewTerm
// returns a Terminal with the name n, which matches a token of the
// lexical category specified by c.
func (t Term) Parse(tokeniser Tokeniser) Result {
	switch token, ok := tokeniser.Peek(); {
	case !ok:
		fallthrough
	case token.category != t.category:
		fallthrough
	case t.exactMatch != "" && token.match != t.exactMatch:
		return NewFailed(t.exactMatch)
	default:
		v, err := t.converter(token.match)
		if err != nil {
			panic(err)
		}
		tokeniser.Inc()
		return NewSucceeded(v, tokeniser.Loc())
	}
}
