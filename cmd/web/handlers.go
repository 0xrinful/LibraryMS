package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/0xrinful/LibraryMS/internal/data"
	"github.com/0xrinful/LibraryMS/internal/validator"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	page := 1
	limit := 4
	offset := (page - 1) * limit

	books, err := app.models.Books.GetBooks(limit, offset)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Books = books
	app.render(w, 200, "home.html", data)
}

func (app *application) profile(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	userID := data.User.ID

	current, err := app.models.BorrowRecord.GetCurrentBorrows(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.CurrentBorrows = current
	data.ActiveBorrows = len(current)

	history, err := app.models.BorrowRecord.GetBorrowHistory(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}
	data.BorrowHistory = history
	data.TotalBorrowed = len(current) + len(history)

	app.render(w, 200, "profile.html", data)
}

func (app *application) dashboard(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, 200, "dashboard.html", data)
}

func (app *application) signup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.DisplayNav = false
	data.Form = userSignupForm{}
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
		data := app.newTemplateData(r)
		data.DisplayNav = false
		data.Form = form
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

type userLoginForm struct {
	Email      string
	Password   string
	RememberMe string
	validator.Validator
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.DisplayNav = false
	data.Form = userLoginForm{}
	app.render(w, 200, "login.html", data)
}

func (app *application) loginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.badRequest(w, r)
		return
	}

	form := userLoginForm{
		Email:      r.FormValue("email"),
		Password:   r.FormValue("password"),
		RememberMe: r.FormValue("remember"),
		Validator:  *validator.New(),
	}

	data.ValidateEmail(&form.Validator, form.Email)
	data.ValidatePasswordPlaintext(&form.Validator, form.Password)

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.DisplayNav = false
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	user, err := app.models.Users.GetByEmail(form.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			form.AddError("email", "User does not exist")
			data := app.newTemplateData(r)
			data.DisplayNav = false
			data.Form = form
			app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		default:
			app.serverError(w, err)
		}
		return
	}

	match, err := user.Password.Matches(form.Password)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if !match {
		form.AddError("password", "Incorrect password")
		data := app.newTemplateData(r)
		data.DisplayNav = false
		data.Form = form
		app.render(w, http.StatusUnprocessableEntity, "login.html", data)
		return
	}

	if form.RememberMe == "1" {
		app.session.Cookie.Persist = true
	} else {
		app.session.Cookie.Persist = false
	}

	err = app.session.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Put(r.Context(), "authenticatedUserID", user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) search(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	app.render(w, 200, "search.html", data)
}

func (app *application) displayBook(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 1 {
		app.notFound(w, r)
		return
	}

	book, err := app.models.Books.GetBookByID(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFound(w, r)
		default:
			app.serverError(w, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Book = book

	app.render(w, http.StatusOK, "book.html", data)
}

func (app *application) borrowBook(w http.ResponseWriter, r *http.Request) {
	bookID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || bookID < 1 {
		app.notFound(w, r)
		return
	}

	userID := app.session.GetInt64(r.Context(), "authenticatedUserID")

	days, err := strconv.Atoi(r.FormValue("days"))
	if err != nil || days < 1 || days > 60 {
		app.flashError(r, "Invalid borrow duration.")
		http.Redirect(w, r, fmt.Sprintf("/books/%d", bookID), http.StatusSeeOther)
		return
	}

	err = app.models.Books.BorrowBook(userID, bookID, days)
	if err != nil {
		switch err {
		case data.ErrAlreadyBorrowed:
			app.flashError(r, "You already borrowed this book.")
		case data.ErrNoAvailableCopies:
			app.flashError(r, "No available copies right now.")
		case data.ErrRecordNotFound:
			app.notFound(w, r)
			return
		default:
			app.serverError(w, err)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/books/%d", bookID), http.StatusSeeOther)
		return
	}

	app.flashInfo(r, "Book borrowed successfully.")
	http.Redirect(w, r, fmt.Sprintf("/books/%d", bookID), http.StatusSeeOther)
}

func (app *application) returnBook(w http.ResponseWriter, r *http.Request) {
	bookID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || bookID < 1 {
		app.notFound(w, r)
		return
	}

	userID := app.session.GetInt64(r.Context(), "authenticatedUserID")

	err = app.models.Books.ReturnBook(userID, bookID)
	if err != nil {
		switch err {
		case data.ErrRecordNotFound:
			app.notFound(w, r)
		default:
			app.serverError(w, err)
		}
		return
	}

	app.flashInfo(r, "Book returned successfully.")
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	err := app.session.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.session.Remove(r.Context(), "authenticatedUserID")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) booksFragment(w http.ResponseWriter, r *http.Request) {
	page := 1
	limit := 4

	if p := r.URL.Query().Get("page"); p != "" {
		page, _ = strconv.Atoi(p)
		if page < 1 {
			page = 1
		}
	}
	offset := (page - 1) * limit

	books, err := app.models.Books.GetBooks(limit, offset)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Books = books
	app.renderPartial(w, "book_cards.html", data)
}
