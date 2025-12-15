package main

import (
	"errors"
	"net/http"

	"github.com/0xrinful/LibraryMS/internal/data"
	"github.com/0xrinful/LibraryMS/internal/validator"
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
		Form:       userSignupForm{},
	}
	app.render(w, 200, "signup.html", data)
}

type userSignupForm struct {
	Name            string
	Email           string
	Password        string
	ConfirmPassword string
	validator.Validator
}

func (app *application) signupPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.badRequest(w, r)
		return
	}

	form := userSignupForm{
		Name:            r.FormValue("name"),
		Email:           r.FormValue("email"),
		Password:        r.FormValue("password"),
		ConfirmPassword: r.FormValue("confirm_password"),
		Validator:       *validator.New(),
	}

	form.Check(validator.NotBlank(form.Name), "name", "must be provided")
	form.Check(len(form.Name) >= 3, "name", "must be more than 3 bytes long")
	form.Check(len(form.Name) <= 500, "name", "must not be more than 500 bytes long")

	data.ValidateEmail(&form.Validator, form.Email)
	data.ValidatePasswordPlaintext(&form.Validator, form.Password)
	if _, ok := form.Errors["password"]; !ok {
		form.Check(
			form.ConfirmPassword == form.Password,
			"confirm_password",
			"passwords do not match",
		)
	}

	if !form.Valid() {
		data := &templateData{DisplayNav: false, Form: form}
		app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	user := &data.User{
		Name:  form.Name,
		Email: form.Email,
	}

	err = user.Password.Set(form.Password)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			form.AddError("email", "this email address already exists")
			data := &templateData{DisplayNav: false, Form: form}
			app.render(w, http.StatusUnprocessableEntity, "signup.html", data)
		default:
			app.serverError(w, err)
		}
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
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
