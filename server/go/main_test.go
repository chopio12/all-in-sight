package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func newTestServer(t *testing.T) *server {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	if err := initDB(db); err != nil {
		t.Fatal(err)
	}
	return &server{db: db}
}

func TestHealthCheck(t *testing.T) {
	app := newTestServer(t)
	recorder := httptest.NewRecorder()

	app.healthCheck(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if body := recorder.Body.String(); body != "{\"status\":\"ok\"}\n" {
		t.Fatalf("body = %q, want health response", body)
	}
}

func TestCreateAndGetHands(t *testing.T) {
	app := newTestServer(t)

	createRecorder := httptest.NewRecorder()
	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/hands",
		strings.NewReader(`{"cards":[{"rank":"A","suit":"s"},{"rank":"K","suit":"h"}]}`),
	)
	app.createHand(createRecorder, createRequest)

	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("POST status = %d, want %d: %s", createRecorder.Code, http.StatusCreated, createRecorder.Body.String())
	}

	listRecorder := httptest.NewRecorder()
	app.getHands(listRecorder, httptest.NewRequest(http.MethodGet, "/hands", nil))

	if listRecorder.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d", listRecorder.Code, http.StatusOK)
	}
	if !strings.Contains(listRecorder.Body.String(), `"rank":"A"`) {
		t.Fatalf("GET body = %q, want stored hand", listRecorder.Body.String())
	}
}

func TestCreateHandRejectsWrongCardCount(t *testing.T) {
	app := newTestServer(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/hands",
		strings.NewReader(`{"cards":[{"rank":"A","suit":"s"}]}`),
	)

	app.createHand(recorder, request)

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnprocessableEntity)
	}
	if !strings.Contains(recorder.Body.String(), "cards must contain exactly 2 cards") {
		t.Fatalf("body = %q, want validation error", recorder.Body.String())
	}
}
