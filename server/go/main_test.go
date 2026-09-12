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
	startActiveSession(t, app)

	createRecorder := httptest.NewRecorder()
	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/hands",
		strings.NewReader(`{"player_id":"player-1","position":"BTN","cards":[{"rank":"A","suit":"s"},{"rank":"K","suit":"h"}]}`),
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
	if !strings.Contains(listRecorder.Body.String(), `"player_id":"player-1"`) || !strings.Contains(listRecorder.Body.String(), `"position":"BTN"`) {
		t.Fatalf("GET body = %q, want player details", listRecorder.Body.String())
	}
	if !strings.Contains(listRecorder.Body.String(), `"session_id":1`) {
		t.Fatalf("GET body = %q, want active session ID", listRecorder.Body.String())
	}
}

func TestGetHandsFiltersBySessionID(t *testing.T) {
	app := newTestServer(t)
	startActiveSession(t, app)
	createHandForTest(t, app, "player-1", "BTN", "A", "s", "K", "h")

	endRecorder := httptest.NewRecorder()
	app.endSession(endRecorder, httptest.NewRequest(http.MethodPost, "/sessions/end", nil))
	if endRecorder.Code != http.StatusOK {
		t.Fatalf("POST end status = %d, want %d", endRecorder.Code, http.StatusOK)
	}
	startActiveSession(t, app)
	createHandForTest(t, app, "player-2", "BB", "Q", "d", "J", "c")

	recorder := httptest.NewRecorder()
	app.getHands(recorder, httptest.NewRequest(http.MethodGet, "/hands?session_id=1", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("GET status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), `"player_id":"player-1"`) || strings.Contains(recorder.Body.String(), `"player_id":"player-2"`) {
		t.Fatalf("GET body = %q, want only hands from session 1", recorder.Body.String())
	}
}

func TestGetHandsRejectsInvalidSessionID(t *testing.T) {
	app := newTestServer(t)
	recorder := httptest.NewRecorder()

	app.getHands(recorder, httptest.NewRequest(http.MethodGet, "/hands?session_id=invalid", nil))

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("GET status = %d, want %d", recorder.Code, http.StatusUnprocessableEntity)
	}
}

func TestCreateHandRejectsWrongCardCount(t *testing.T) {
	app := newTestServer(t)
	startActiveSession(t, app)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/hands",
		strings.NewReader(`{"player_id":"player-1","position":"BTN","cards":[{"rank":"A","suit":"s"}]}`),
	)

	app.createHand(recorder, request)

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnprocessableEntity)
	}
	if !strings.Contains(recorder.Body.String(), "cards must contain exactly 2 cards") {
		t.Fatalf("body = %q, want validation error", recorder.Body.String())
	}
}

func TestCreateHandRejectsDuplicateCardsInOneHand(t *testing.T) {
	app := newTestServer(t)
	startActiveSession(t, app)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/hands",
		strings.NewReader(`{"player_id":"player-1","position":"BTN","cards":[{"rank":"A","suit":"s"},{"rank":"A","suit":"s"}]}`),
	)

	app.createHand(recorder, request)

	if recorder.Code != http.StatusUnprocessableEntity {
		t.Fatalf("POST status = %d, want %d: %s", recorder.Code, http.StatusUnprocessableEntity, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "cards must not contain the same card twice") {
		t.Fatalf("POST body = %q, want duplicate card error", recorder.Body.String())
	}
}

func TestCreateHandRejectsCardUsedInSameSession(t *testing.T) {
	app := newTestServer(t)
	startActiveSession(t, app)
	createHandForTest(t, app, "player-1", "BTN", "A", "s", "K", "h")

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/hands",
		strings.NewReader(`{"player_id":"player-2","position":"BB","cards":[{"rank":"A","suit":"s"},{"rank":"Q","suit":"d"}]}`),
	)
	app.createHand(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("POST status = %d, want %d: %s", recorder.Code, http.StatusConflict, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), "a card is already used in this session") {
		t.Fatalf("POST body = %q, want duplicate card error", recorder.Body.String())
	}
}

func TestHandsRequireActiveSession(t *testing.T) {
	app := newTestServer(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/hands",
		strings.NewReader(`{"player_id":"player-1","position":"BTN","cards":[{"rank":"A","suit":"s"},{"rank":"K","suit":"h"}]}`),
	)

	app.createHand(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
	}
}

func TestSessionLifecycle(t *testing.T) {
	app := newTestServer(t)
	startActiveSession(t, app)
	duplicateStartRecorder := httptest.NewRecorder()
	app.startSession(duplicateStartRecorder, httptest.NewRequest(http.MethodPost, "/sessions/start", nil))
	if duplicateStartRecorder.Code != http.StatusConflict {
		t.Fatalf("duplicate POST start status = %d, want %d", duplicateStartRecorder.Code, http.StatusConflict)
	}

	currentRecorder := httptest.NewRecorder()
	app.getCurrentSession(currentRecorder, httptest.NewRequest(http.MethodGet, "/sessions/current", nil))
	if currentRecorder.Code != http.StatusOK {
		t.Fatalf("GET current status = %d, want %d", currentRecorder.Code, http.StatusOK)
	}

	endRecorder := httptest.NewRecorder()
	app.endSession(endRecorder, httptest.NewRequest(http.MethodPost, "/sessions/end", nil))
	if endRecorder.Code != http.StatusOK {
		t.Fatalf("POST end status = %d, want %d", endRecorder.Code, http.StatusOK)
	}

	noCurrentRecorder := httptest.NewRecorder()
	app.getCurrentSession(noCurrentRecorder, httptest.NewRequest(http.MethodGet, "/sessions/current", nil))
	if noCurrentRecorder.Code != http.StatusNotFound {
		t.Fatalf("GET current after end status = %d, want %d", noCurrentRecorder.Code, http.StatusNotFound)
	}
}

func startActiveSession(t *testing.T, app *server) {
	t.Helper()
	recorder := httptest.NewRecorder()
	app.startSession(recorder, httptest.NewRequest(http.MethodPost, "/sessions/start", nil))
	if recorder.Code != http.StatusCreated {
		t.Fatalf("POST start status = %d, want %d: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
}

func createHandForTest(t *testing.T, app *server, playerID, position, firstRank, firstSuit, secondRank, secondSuit string) {
	t.Helper()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/hands",
		strings.NewReader(`{"player_id":"`+playerID+`","position":"`+position+`","cards":[{"rank":"`+firstRank+`","suit":"`+firstSuit+`"},{"rank":"`+secondRank+`","suit":"`+secondSuit+`"}]}`),
	)
	app.createHand(recorder, request)
	if recorder.Code != http.StatusCreated {
		t.Fatalf("POST status = %d, want %d: %s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
}
