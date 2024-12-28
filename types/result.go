package types

type None struct{}

// locs is a map representing a set of unique locations or indicies into
// a token at which a parser succesfully finished recognising a sequence
// of tokens. A location set of length 0 implies that the paser failed.
// A location set of length > 1 implies ambiguity
type locs map[int]None

func NewLocs(l int) locs {
	return locs{
		l: None{},
	}
}

// Merge performs the Union of locations sets a and b. The meaning of
// the returned value is a set of locations for which a parser succeeded
func (a locs) Merge(b locs) locs {
	for l := range b {
		a[l] = None{}
	}
	return a
}

// parseError represents an error encountered by a parser, or a reason
// or indicator for a failed parse result.
type parseError struct {
	// expected indicates that a parser failed because
	// it encountered an unexpected token sequence, or
	// one different to the one it was expecting
	expected string
}

func newParseError(s string) parseError {
	return parseError{s}
}

// Result represents the result of applying a parser to an input text The
// locations or Locs returns by Locs() and the parse errors or ParseErrors
// returned by Errors() are mutually exclusive; that is, if one is
// not-empty, the other should be empty. This is because a Parser either
// succeedes or fails. If a parser succeeds, it returns a succesfful
// result containing the locations in the token stream where the parser
// successfull finished recognising the sequence and a Value which is user
// determined. If parser fails it returns a result containing the reasons
// for why
type Result interface {
	// merge combines another parse result together with
	// "this" one
	merge(Result) Result

	// Join combines two parse results according to some
	// meaningful predicate
	Join(Result) Result

	// Locs returns a set of locations representing the
	// locations at which a paser successfully finished
	// recognising a sequence of tokens. If a Result
	// contains a non-empty list of locations as
	// returned by Locs(), it should contain an empty
	// list of errors as returned by Errors()
	Locs() locs

	// value is the user determined value returned as
	// part of a successful parse result, or the value
	// ultimately propagated back. The interpretation of
	// this value is defined by the user, this could
	// could be anything, e.g. an integer representing
	// the result of an arithmetic expression, or an
	// abstract syntax tree representing source code
	value() any

	// Errors returns a list of errors or reasons for
	// why the parser failed. If a Result contains a
	// non-empty list of erors as returned by Errors(),
	// it should contain an empty list of locs as
	// returnd by Locs()
	Errors() []parseError
}

// Succeeded implements the Result interface for a "successful" parse
// result. Returning a Succeeded means the parser successfully finished
// recognising a sequence of tokens at the locations stored in locs
type Succeeded struct {
	// locs is a set of locations at which a paser
	// successfully finished recognising a sequence of
	// tokens. This will always be non-empty
	locs locs

	// v is the user determined v returned by Value()
	v any
}

func NewSucceeded(s any, l int) Result {
	return Succeeded{NewLocs(l), s}
}

// merge combines the Succeeded parse results a and b and their location
// sets and internal values
func (a Succeeded) merge(b Result) Result {
	r := b.(Succeeded)
	a.locs = a.locs.Merge(r.locs)
	a.v = r.v
	return a
}

// Locs returns a set of locations representing the locations at which a
// paser successfully finished recognising a sequence of tokens. For a
// Succeeded result, the returned set will always be non-empty
func (s Succeeded) Locs() locs {
	return s.locs
}

// value returns a user defined value returned as the result of a
// successful execution of a parser
func (s Succeeded) value() any {
	return s.v
}

// Errors returns a list of errors or reasons for why the parser failed.
// for A succueeded Result, the returned list will always be empty
func (Succeeded) Errors() []parseError {
	return nil
}

// Join joins the result b with the result a. This is just the result of
// merging a and b if b is also a Succeeded result, or just a if b is a
// Failed result
func (a Succeeded) Join(b Result) Result {
	switch b.(type) {
	case Succeeded:
		return a.merge(b)
	case Failed:
		return a
	}
	panic(ErrInternal)
}

// Failed implements the Result interface for a "failed" parse
// result. Returning a Failed means the parser failed to reecognise the
// applied token sequence
type Failed struct {
	// parseErrors is a list of errors or reasons for
	// why the parser failed. This iwll always be non-empty
	parseErrors []parseError
}

func NewFailed(s string) Result {
	return Failed{
		parseErrors: []parseError{
			newParseError(s),
		},
	}
}

// merge combines the Failed parse results a and b by merging their
// parse errors
func (a Failed) merge(b Result) Result {
	r := b.(Failed)
	a.parseErrors = append(a.parseErrors, r.Errors()...)
	return a
}

// Locs usually returns a set of locations representing the locations at
// which a paser successfully finished recognising a sequence of tokens.
// In this case the set will always be empty
func (Failed) Locs() locs {
	return locs{}
}

// value usually returns a user defined value returned as the result of a
// successful execution of a parser. This will always be nil in this
// case
func (Failed) value() any {
	return nil
}

// Errors returns a list of errors or reasons for why the parser failed.
// for A failed result, this will always be non-empty
func (f Failed) Errors() []parseError {
	return f.parseErrors
}

// Join joins the result b with the result a. This is just the result of
// merging a and b if b is also a Failed result, or just a if b is a
// Succeed result
func (a Failed) Join(b Result) Result {
	switch b.(type) {
	case Succeeded:
		return b
	case Failed:
		return a.merge(b)
	}
	panic(ErrInternal)
}
