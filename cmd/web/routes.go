package main

import (
	"net/http"

	"github.com/0xrinful/rush"

	"github.com/0xrinful/LibraryMS/ui"
)

func (app *application) routes() http.Handler {
	r := rush.New()
	r.NotFound = http.HandlerFunc(app.notFound)

	fileServer := http.FileServer(http.FS(ui.Files))
	r.Handle("/static/*", fileServer, "GET")

	r.Get("/", app.home)
	r.Get("/profile", app.profile)
	r.Get("/dashboard", app.dashboard)

	r.Get("/signup", app.signup)
	r.Post("/signup", app.signupPost)

	r.Get("/login", app.login)
	r.Get("/search", app.search)
	r.Get("/book", app.displayBook)

	return r
}
