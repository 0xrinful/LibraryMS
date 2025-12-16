package main

import (
	"net/http"

	"github.com/0xrinful/rush"

	"github.com/0xrinful/LibraryMS/ui"
)

func (app *application) routes() http.Handler {
	r := rush.New()
	r.Use(app.session.LoadAndSave, app.authenticate)

	r.NotFound = http.HandlerFunc(app.notFound)

	fileServer := http.FileServer(http.FS(ui.Files))
	r.Handle("/static/*", fileServer, "GET")

	r.Get("/", app.home)

	r.Group(func(r *rush.Router) {
		r.Use(app.requireNoAuthentication)
		r.Get("/signup", app.signup)
		r.Post("/signup", app.signupPost)
		r.Get("/login", app.login)
		r.Post("/login", app.loginPost)
	})
	r.Get("/logout", app.logout)

	r.Get("/search", app.search)
	r.Get("/book/{id}", app.displayBook)
	r.Get("/books", app.booksFragment)

	r.Group(func(r *rush.Router) {
		r.Use(app.requireAuthentication)

		r.Get("/profile", app.profile)
		r.Get("/dashboard", app.dashboard)
	})

	return r
}
