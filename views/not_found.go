package views

import (
	"net/http"
)

var notFoundTmpl = renderer.GetTemplate("404.html")

func NotFound() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		notFoundTmpl.ViewData(w, r, "Not Found").
			WithNavItems(defaultNavItems).
			Render(404)
	}
}