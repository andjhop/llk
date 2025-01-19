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
	// Expected indicates that a parser failed because
	// it encountered an unexpected token sequence, or
	// one different to the one it was expecting
	Expected string

	// Found is the lexical item that was actually
	// encountered instead of the one we expected
	Found string

	// Line and Column reflect the line number and
	// column number of the input text where the error
	// occured
	Line, Column int
}

func newParseError(expected, found string) parseError {
	return parseError{expected, found, 0, 0}
}

// WithLineAndColumn returns
func (e parseError) WithLineAndColumn(line, column int) parseError {
	e.Line = line
	e.Column = column
	return e
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
type Result[V any] interface {
	// merge combines another parse result together with
	// "this" one
	merge(Result[V]) Result[V]

	// Join combines two parse results according to some
	// meaningful predicate
	Join(Result[V]) Result[V]

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
	value() V

	// Errors returns a list of errors or reasons for
	// why the parser failed. If a Result contains a
	// non-empty list of erors as returned by Errors(),
	// it should contain an empty list of locs as
	// returnd by Locs()
	Errors() []parseError
}

// Halt is a special kind of Result that should never be handled.
// Returning a Halt as a parser result is a signal to halt execution of
// the parser.
type Halt[V any] struct {
	Result[V]

	// component helps to categorise the result
	component string

	// Message is a Message explaining why this result
	// was returned
	Message string

	Line, Column int
}

func NewHalt[V any](c, m string) Halt[V] {
	return Halt{
		component: c,
		Message:   m,
	}
}

func (h Halt[V]) WithLineAndColumn(line, column int) Halt[V] {
	h.Line = line
	h.Column = column
	return h
}

// Succeeded implements the Result interface for a "successful" parse
// result. Returning a Succeeded means the parser successfully finished
// recognising a sequence of tokens at the locations stored in locs
type Succeeded[V any] struct {
	// locs is a set of locations at which a paser
	// successfully finished recognising a sequence of
	// tokens. This will always be non-empty
	locs locs

	// v is the user determined v returned by Value()
	v V
}

func NewSucceeded[V any](s any, l int) Result[V] {
	return Succeeded{NewLocs(l), s}
}

// merge combines the Succeeded parse results a and b and their location
// sets and internal values
func (a Succeeded[V]) merge(b Result[V]) Result[V] {
	r := b.(Succeeded)
	a.locs = a.locs.Merge(r.locs)
	a.v = r.v
	return a
}

// Locs returns a set of locations representing the locations at which a
// paser successfully finished recognising a sequence of tokens. For a
// Succeeded result, the returned set will always be non-empty
func (s Succeeded[V]) Locs() locs {
	return s.locs
}

// value returns a user defined value returned as the result of a
// successful execution of a parser
func (s Succeeded[V]) value() V {
	return s.v
}

// Errors returns a list of errors or reasons for why the parser failed.
// for A succueeded Result, the returned list will always be empty
func (Succeeded[V]) Errors() []parseError {
	return nil
}

// Join joins the result b with the result a. This is just the result of
// merging a and b if b is also a Succeeded result, or just a if b is a
// Failed result
func (a Succeeded[V]) Join(b Result[V]) Result[V] {
	switch b.(type) {
	case Succeeded:
		return a.merge(b)
	case Failed:
		return a
	case Halt:
		return b
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

func NewFailed[V any](expected, found string) Result[V] {
	return Failed{
		parseErrors: []parseError{
			newParseError(expected, found),
		},
	}
}

func NewScanFailed[V any](expected, found string, line, col int) Result[V] {
	return Failed{
		parseErrors: []parseError{
			newParseError(expected, found).
				WithLineAndColumn(line, col),
		},
	}
}

// merge combines the Failed parse results a and b by merging their
// parse errors
func (a Failed) merge(b Result[any]) Result[any] {
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
func (a Failed) Join(b Result[any]) Result[any] {
	switch b.(type) {
	case Succeeded:
		return b
	case Failed:
		return a.merge(b)
	case Halt:
		return b
	}
	panic(ErrInternal)
}
