package main

import (
	"html/template"
	"io/fs"
	"path/filepath"

	"github.com/0xrinful/LibraryMS/internal/data"
	"github.com/0xrinful/LibraryMS/ui"
)

type templateData struct {
	FlashInfo       string
	FlashError      string
	DisplayNav      bool
	Form            any
	IsAuthenticated bool
	User            *data.User
	Book            *data.Book
	Books           []*data.Book

	CurrentBorrows []*data.BorrowedBook
	ActiveBorrows  int
	TotalBorrowed  int
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(ui.Files, "html/pages/*.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/base.html",
			"html/partials/*.html",
			page,
		}

		ts, err := template.New(name).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	partials, err := fs.Glob(ui.Files, "html/partials/*.html")
	if err != nil {
		return nil, err
	}

	for _, partial := range partials {
		name := filepath.Base(partial)
		ts, err := template.New(name).ParseFS(ui.Files, partial)
		if err != nil {
			return nil, err
		}
		cache[name] = ts
	}

	return cache, nil
}
