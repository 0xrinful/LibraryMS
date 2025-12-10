package main

import (
	"net/http"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := &templateData{
		DisplayNav: true,
	}
	app.render(w, 200, "home.html", data)
}

func (app *application) profile(w http.ResponseWriter, r *http.Request) {
	data := &templateData{
		DisplayNav: true,
	}
	app.render(w, 200, "profile.html", data)
}

func (app *application) dashboard(w http.ResponseWriter, r *http.Request) {
	data := &templateData{
		DisplayNav: true,
	}
	app.render(w, 200, "dashboard.html", data)
}

func (app *application) signup(w http.ResponseWriter, r *http.Request) {
	data := &templateData{
		DisplayNav: false,
	}
	app.render(w, 200, "signup.html", data)
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	data := &templateData{
		DisplayNav: false,
	}
	app.render(w, 200, "login.html", data)
}

func (app *application) search(w http.ResponseWriter, r *http.Request) {
	data := &templateData{
		DisplayNav: true,
	}
	app.render(w, 200, "search.html", data)
}

func (app *application) displayBook(w http.ResponseWriter, r *http.Request) {
	data := &templateData{
		DisplayNav: true,
	}
	app.render(w, 200, "book.html", data)
}
