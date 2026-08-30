package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// parseIDParam extracts and parses an int64 URL parameter, writing a 400
// response and returning ok=false if it's missing or invalid.
func parseIDParam(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	raw := chi.URLParam(r, name)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid "+name)
		return 0, false
	}
	return id, true
}
