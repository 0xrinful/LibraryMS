package main

import (
	"fmt"
	"net/http"
	"path/filepath"

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

func (app *application) renderPartial(
	w http.ResponseWriter,
	templateName string,
	data *templateData,
) {
	ts, ok := app.templateCache[templateName]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", templateName)
		app.serverError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	name := templateName[:len(templateName)-len(filepath.Ext(templateName))]
	err := ts.ExecuteTemplate(w, name, data)
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
		FlashInfo:       app.session.PopString(r.Context(), "flash_info"),
		FlashError:      app.session.PopString(r.Context(), "flash_error"),
	}

	user, ok := r.Context().Value(userContextKey).(*data.User)
	if ok {
		td.User = user
	}

	return td
}

func (app *application) flashInfo(r *http.Request, msg string) {
	app.session.Put(r.Context(), "flash_info", msg)
}

func (app *application) flashError(r *http.Request, msg string) {
	app.session.Put(r.Context(), "flash_error", msg)
}
