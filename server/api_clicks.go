// Package server is boop
package server

import (
	"net/http"
)

// Click reports clicks for the given team
func Click(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
}
