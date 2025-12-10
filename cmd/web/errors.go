package main

import (
	"fmt"
	"net/http"
)

func (app *application) serverError(w http.ResponseWriter, err error) {
	app.logger.PrintError(err)
	w.WriteHeader(500)

	page := "500.html"
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.logger.PrintError(err)
		return
	}

	err = ts.ExecuteTemplate(w, "base", templateData{DisplayNav: false})
	if err != nil {
		app.logger.PrintError(err)
	}
}

func (app *application) notFound(w http.ResponseWriter, r *http.Request) {
	app.render(w, 404, "404.html", &templateData{DisplayNav: false})
}
