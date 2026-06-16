package server

import "net/http"

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// handleGraphQLNotImplemented is the authenticated mount point for the GraphQL
// API. The schema and resolvers arrive in a later issue; this placeholder keeps
// the auth middleware wired and exercisable in the meantime.
func handleGraphQLNotImplemented(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "graphql api not yet implemented", http.StatusNotImplemented)
}
