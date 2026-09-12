package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

func (app *server) createHand(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var request handRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "invalid JSON request body")
		return
	}

	if request.PlayerID == "" {
		writeError(w, http.StatusUnprocessableEntity, "player_id is required")
		return
	}
	if request.Position == "" {
		writeError(w, http.StatusUnprocessableEntity, "position is required")
		return
	}
	if err := validateCards(request.Cards); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	sessionID, err := app.getActiveSessionID()
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusConflict, "no active session")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load active session")
		return
	}
	if err := app.validateCardsUnusedInSession(sessionID, request.Cards); err != nil {
		if errors.Is(err, errCardAlreadyUsed) {
			writeError(w, http.StatusConflict, "a card is already used in this session")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to validate cards")
		return
	}

	createdAt := time.Now().UTC().Format(time.RFC3339)
	result, err := app.db.Exec(
		"INSERT INTO hands (session_id, player_id, position, hand_number_1, hand_number_2, hand_sute_1, hand_sute_2, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		sessionID, request.PlayerID, request.Position, request.Cards[0].Rank, request.Cards[1].Rank, request.Cards[0].Suit, request.Cards[1].Suit, createdAt,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save hand")
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get hand ID")
		return
	}
	writeJSON(w, http.StatusCreated, handResponse{
		ID:        id,
		SessionID: sessionID,
		PlayerID:  request.PlayerID,
		Position:  request.Position,
		Cards:     request.Cards,
		CreatedAt: createdAt,
	})
}

func (app *server) getHands(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT id, session_id, player_id, position, hand_number_1, hand_number_2, hand_sute_1, hand_sute_2, created_at
		FROM hands
	`
	var queryArgs []any
	if sessionIDValue := r.URL.Query().Get("session_id"); sessionIDValue != "" {
		sessionID, err := strconv.ParseInt(sessionIDValue, 10, 64)
		if err != nil || sessionID < 1 {
			writeError(w, http.StatusUnprocessableEntity, "session_id must be a positive integer")
			return
		}
		query += "WHERE session_id = ?\n"
		queryArgs = append(queryArgs, sessionID)
	}
	query += "ORDER BY id DESC"
	rows, err := app.db.Query(query, queryArgs...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load hands")
		return
	}
	defer rows.Close()

	hands := make([]handResponse, 0)
	for rows.Next() {
		var hand handResponse
		var firstRank, secondRank, firstSuit, secondSuit string
		if err := rows.Scan(&hand.ID, &hand.SessionID, &hand.PlayerID, &hand.Position, &firstRank, &secondRank, &firstSuit, &secondSuit, &hand.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read hand")
			return
		}
		hand.Cards = []card{{Rank: firstRank, Suit: firstSuit}, {Rank: secondRank, Suit: secondSuit}}
		hands = append(hands, hand)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load hands")
		return
	}
	writeJSON(w, http.StatusOK, hands)
}
