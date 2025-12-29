package main

type contextKey string

const (
	isAuthenticatedContextKey = contextKey("isAuthenticated")
	nipContextKey = contextKey("companyNIP")
)
