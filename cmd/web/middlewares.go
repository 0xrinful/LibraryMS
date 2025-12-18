package main

import (
	"context"
	"net/http"

	"github.com/0xrinful/LibraryMS/internal/data"
)

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := app.session.GetInt64(ctx, "authenticatedUserID")
		if userID == 0 {
			ctx = context.WithValue(ctx, isAuthenticatedContextKey, false)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		user, err := app.models.Users.Get(userID)
		if err != nil {
			app.session.Remove(ctx, "authenticatedUserID")
			ctx = context.WithValue(ctx, isAuthenticatedContextKey, false)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ctx = context.WithValue(ctx, isAuthenticatedContextKey, true)
		ctx = context.WithValue(ctx, userContextKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(userContextKey).(*data.User)
		if !ok || user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireNoAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(userContextKey).(*data.User)
		if ok && user != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
