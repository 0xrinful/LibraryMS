package main

type contextKey string

const (
	isAuthenticatedContextKey contextKey = "isAuthenticated"
	userContextKey            contextKey = "user"
)
