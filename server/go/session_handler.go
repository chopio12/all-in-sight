package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"
)

func (app *server) healthCheck(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (app *server) startSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	createdAt := time.Now().UTC().Format(time.RFC3339)
	result, err := app.db.Exec(
		"INSERT INTO poker_sessions (status, created_at) VALUES ('active', ?)",
		createdAt,
	)
	if err != nil {
		writeError(w, http.StatusConflict, "an active session already exists")
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get session ID")
		return
	}
	writeJSON(w, http.StatusCreated, sessionResponse{ID: id, Status: "active", CreatedAt: createdAt})
}

func (app *server) endSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	result, err := app.db.Exec("UPDATE poker_sessions SET status = 'closed' WHERE status = 'active'")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to end session")
		return
	}
	updated, err := result.RowsAffected()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to end session")
		return
	}
	if updated == 0 {
		writeError(w, http.StatusNotFound, "no active session")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "closed"})
}

func (app *server) getCurrentSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var session sessionResponse
	err := app.db.QueryRow(`
		SELECT id, status, created_at
		FROM poker_sessions
		WHERE status = 'active'
	`).Scan(&session.ID, &session.Status, &session.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "no active session")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load active session")
		return
	}
	writeJSON(w, http.StatusOK, session)
}
