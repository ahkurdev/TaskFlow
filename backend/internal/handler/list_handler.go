package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"taskflow/internal/middleware"
	"taskflow/internal/service"
)

type ListHandler struct {
	lists *service.ListService
}

func NewListHandler(lists *service.ListService) *ListHandler {
	return &ListHandler{lists: lists}
}

type listRequest struct {
	Name string `json:"name"`
}

type reorderRequest struct {
	Position int `json:"position"`
}

func (h *ListHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	boardID, ok := parseIDParam(w, r, "boardID")
	if !ok {
		return
	}

	var req listRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	list, err := h.lists.Create(r.Context(), userID, boardID, req.Name)
	if err != nil {
		writeListError(w, err, "failed to create list")
		return
	}

	writeJSON(w, http.StatusCreated, list)
}

func (h *ListHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	boardID, ok := parseIDParam(w, r, "boardID")
	if !ok {
		return
	}
	listID, ok := parseIDParam(w, r, "listID")
	if !ok {
		return
	}

	var req listRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	list, err := h.lists.Update(r.Context(), userID, boardID, listID, req.Name)
	if err != nil {
		writeListError(w, err, "failed to update list")
		return
	}

	writeJSON(w, http.StatusOK, list)
}

func (h *ListHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	boardID, ok := parseIDParam(w, r, "boardID")
	if !ok {
		return
	}
	listID, ok := parseIDParam(w, r, "listID")
	if !ok {
		return
	}

	var req reorderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.lists.Reorder(r.Context(), userID, boardID, listID, req.Position); err != nil {
		writeListError(w, err, "failed to reorder list")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ListHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	boardID, ok := parseIDParam(w, r, "boardID")
	if !ok {
		return
	}
	listID, ok := parseIDParam(w, r, "listID")
	if !ok {
		return
	}

	if err := h.lists.Delete(r.Context(), userID, boardID, listID); err != nil {
		writeListError(w, err, "failed to delete list")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeListError(w http.ResponseWriter, err error, fallback string) {
	if errors.Is(err, service.ErrNotFound) {
		writeError(w, http.StatusNotFound, "list or board not found")
		return
	}
	writeError(w, http.StatusInternalServerError, fallback)
}
