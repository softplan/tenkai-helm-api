package handlers

import (
	"net/http"
)

func (appContext *AppContext) health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
