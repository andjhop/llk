package types

import (
	"errors"
)

var (
	// ErrInternal is an error indicating that
	// something unexpected went wrong internally
	ErrInternal = errors.New("internal")

	// ErrInvalidLoc is an internal error which
	// indicates the scanner hit an invalid location
	ErrBadLoc = errors.New("invalid location")

	// ErrBadCharset indicates the tokeniser scanned a
	// a character belonging to an unsupported
	// character set
	ErrBadCharset = errors.New("invalid charset")

	// ErrEOF indicates that the tokeniser reached the
	// end of the input
	ErrEOF = errors.New("end of file")
)
