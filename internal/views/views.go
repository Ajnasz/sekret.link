package views

import "net/http"

type View[T any] interface {
	Render(w http.ResponseWriter, r *http.Request, data T)
	RenderError(w http.ResponseWriter, r *http.Request, err error)
}
