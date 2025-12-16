package main

import (
	"fmt"
	"net/http"

	"github.com/0xrinful/LibraryMS/internal/data"
)

func (app *application) render(w http.ResponseWriter, status int, page string, data *templateData) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	w.WriteHeader(status)

	err := ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, err)
	}
}

func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	return ok && isAuthenticated
}

func (app *application) newTemplateData(r *http.Request) *templateData {
	td := &templateData{
		IsAuthenticated: app.isAuthenticated(r),
		DisplayNav:      true,
	}

	user, ok := r.Context().Value(userContextKey).(*data.User)
	if ok {
		td.User = user
	}

	return td
}
