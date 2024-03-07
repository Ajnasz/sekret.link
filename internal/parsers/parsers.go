package parsers

import "net/http"

type Parser[T any] interface {
	Parse(r *http.Request) (T, error)
}
