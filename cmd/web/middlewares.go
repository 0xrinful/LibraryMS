package main

import (
	"context"
	"net/http"
)

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		userID := app.session.GetInt(ctx, "authenticatedUserID")
		if userID == 0 {
			ctx = context.WithValue(ctx, isAuthenticatedContextKey, false)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		user, err := app.models.Users.Get(userID)
		if err != nil {
			// stale / invalid session → logout silently
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
		isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
		if !ok || !isAuthenticated {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) requireNoAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
		if ok && isAuthenticated {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
