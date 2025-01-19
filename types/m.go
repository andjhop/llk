package types

// lazy represents a continution which takes a Result r; the result of
// the "previous" parse in a chain. This can be used to delay the choice
// of the the "next" parser until parse time, and is useful for defining
// parsers recursively
type lazy[V any] func(r V) Parser[V]

func NewLazy[V any](p Parser[V]) lazy[V] {
	return func(V) Parser[V] {
		return p
	}
}

func Wrap[V any](f func(V) V) lazy[V] {
	return func(r V) Parser[V] {
		return NewEmpty[V](f(r))
	}
}

// M is a "chainable" Parser, or a chain of continuations where each
// continuation is applied to the result of the previous. This is useful
// for building combinators which combine the results of multiple
// parsers. The previous result is stored in result, the folder function
// takes the previous result and next continuation and combines them to
// create the next result. How parsers are chained together is
// determined by the behvaiour of the folder function. Chaining
// continuations together should be done in a way that the chaining
// operation is associative
type M[V any] struct {
	// name is the optional name of the parser used in
	// error messages
	name string

	// folder controls how continuation are chanied
	// together, it takes the previous result and next
	// continuation and combines them to create the next
	// result
	folder func(*M[V], Tokeniser) Result[V]

	// result is the result of invoking the previous
	// continuation is the result of invoking the
	// previous continuation
	result Result[V]

	// lazies is the next continuation result is the
	// result of invoking the previous continuation
	// right is the next continuation result is the
	// result of invoking the
	lazies []lazy[V]
}

func NewM[V any](f func(*M[V], Tokeniser) Result[V]) *M[V] {
	return &M[V]{
		folder: f,
		result: Failed{},
	}
}

// WithName sets the name of M to the name given by n, this is the name
// used in parser error messages
func (m *M[V]) WithName(s string) *M[V] {
	m.name = s
	return m
}

// WithResult sets the value of the of the "previous" continuation
// result to r, this represents a continuation who
func (m *M[V]) WithResult(r Result[V]) *M[V] {
	m.result = r
	return m
}

// WithLazies returns a chainable parser with the continations
// specified by lazies
func (m *M[V]) WithLazies(lazies ...lazy[V]) *M[V] {
	m.lazies = append(m.lazies, lazies...)
	return m
}

func (m *M[V]) Name() string {
	return m.name
}

// Result returns the result of invoking the previous continuation is
// the result of invoking the previous continuation
func (m *M[V]) Result() Result[V] {
	return m.result
}

// Passthrough chains a parser on to the end of m return the result of
// the previous continuation, this is short hand for chaining
// continuations which are "transparent" as far as computing a result
// value, so:
//
//	// tbc
//
// ...
func (m *M[V]) Passthrough(p Parser[V]) *M[V] {
	n := NewM(m.folder).Chain(p)
	return m.Lazy(func(v any) Parser[V] {
		return n.Return(func(any) any {
			return v
		})
	})
}

// Text is shorthand for chaining a text parser onto m, it is equivalent
// to:
//
//	m.Passthrough(Text(category))
func (m *M[V]) Text(category rune) *M[V] {
	return m.Passthrough(Text(category))
}

// Id is shorthand for chaining an ident parser onto m, it is equivalent
// to:
//
//	m.Passthrough(Id(s))
func (m *M[V]) Id(s string) *M[V] {
	return m.Passthrough(Id(s))
}

// Int is shorthand for chaining an integer parser onto m, it is
// equivalent to:
//
//	m.Passthrough(Int())
func (m *M[V]) Int() *M[V] {
	return m.Passthrough(Int())
}

// Float is shorthand for chaining a flaot parser onto m, it is
// equivalent to:
//
//	m.Passthrough(Float)
func (m *M[V]) Float() *M[V] {
	return m.Passthrough(Float())
}

// String is shorthand for chaining a string parser onto m, it is
// equivalent to:
//
//	m.Passthrough(String())
func (m *M[V]) String() *M[V] {
	return m.Passthrough(String())
}

// Lazy chains a continuation which chooses the "next" Parser to
// continue the execution with on to the end of m, this will be invoked
// with the result of the "previous" continuation. How parsers are
// chained together is determined by the behvaiour of the folder
// function. Chaining continuations together should be done in a way
// that the chaining operation is associative, that is:
//
//	a.Lazy(func (any) Parser {
//		return b.Lazy(func(any) Parser {
//			return c
//		})
//	})
//
//  Should return a Parser which executes equivalently to:
//
//		a.Lazy(func(any) Parser {
//			return b
//		}).Lazy(func(any) Parser {
//			return c
//		})

func (m *M[V]) Lazy(lazies ...lazy[V]) *M[V] {
	return NewM(m.folder).
		WithResult(m.result).
		WithLazies(m.lazies...).
		WithLazies(lazies...)
}

// Chain chains a parser on to the end of m, this is just shorthand for
// chaining a parser onto m which doesn't care about the "previous"
// result, so:
//
//	a.Lazy(func (any) Parser {
//		return b
//	})
//
// Can be abbreviated to:
//
//	a.Chain(b)
func (m *M[V]) Chain(p Parser[V]) *M[V] {
	return m.Lazy(NewLazy(p))
}

// Return chains a parser on to the end of m, this is just shorthand for
// calling Lazy with with a continuation which returns an empty. That
// is, Return is useful for extacting the value of a successful parse
// result where parsing "ends", so:
//
//	a.Lazy(func (v any) Parser {
//		return func(v any) Parser {
//			return NewEmpty(f(v))
//		}
//	})
//
// Can be abbreviated to:
//
//	a.Return(f)
func (m *M[V]) Return(f func(any) any) *M[V] {
	return m.Lazy(Wrap(f))
}

// Parse invokes a folder function to combine continuations in the
// chain. A folder is called with a continuation b and the with the
// result obtained from applying the parser returned by continuation a
// to the token stream. The result returned by the folder function over
// the continuation chain is the parse result.
func (m *M[V]) Parse(t Tokeniser) (r Result[V]) {
	if len(m.lazies) == 0 {
		return
	}
	lazy := m.lazies[0]
	r = lazy(m.result.value()).Parse(t)
	if halt, ok := r.(Halt[any]); ok {
		return halt
	}

	if len(m.lazies) == 1 {
		return
	}
	lazies := m.lazies[1:]
	r = m.folder(NewM(m.folder).
		WithName(m.name).
		WithResult(r).
		WithLazies(lazies...), t)
	return
}
