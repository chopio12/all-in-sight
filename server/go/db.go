package main

import (
	"database/sql"
	"errors"
)

func initDB(db *sql.DB) error {
	_, err := db.Exec(`
		PRAGMA foreign_keys = ON;

		CREATE TABLE IF NOT EXISTS poker_sessions (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			status     TEXT    NOT NULL CHECK (status IN ('active', 'closed')),
			created_at TEXT    NOT NULL DEFAULT (datetime('now'))
		);

		CREATE UNIQUE INDEX IF NOT EXISTS one_active_poker_session
		ON poker_sessions (status)
		WHERE status = 'active';

		CREATE TABLE IF NOT EXISTS hands (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id    INTEGER NOT NULL,
			player_id     TEXT    NOT NULL,
			position      TEXT    NOT NULL,
			hand_number_1 TEXT    NOT NULL,
			hand_number_2 TEXT    NOT NULL,
			hand_sute_1   TEXT    NOT NULL,
			hand_sute_2   TEXT    NOT NULL,
			created_at    TEXT    NOT NULL DEFAULT (datetime('now')),
			FOREIGN KEY (session_id) REFERENCES poker_sessions(id),
			UNIQUE (session_id, position)
		)
	`)
	return err
}

func (app *server) getActiveSessionID() (int64, error) {
	var sessionID int64
	err := app.db.QueryRow("SELECT id FROM poker_sessions WHERE status = 'active'").Scan(&sessionID)
	return sessionID, err
}

var errCardAlreadyUsed = errors.New("card already used in session")

func (app *server) validateCardsUnusedInSession(sessionID int64, cards []card) error {
	for _, card := range cards {
		var exists bool
		err := app.db.QueryRow(`
			SELECT EXISTS (
				SELECT 1
				FROM hands
				WHERE session_id = ?
				  AND (
					(hand_number_1 = ? AND hand_sute_1 = ?)
					OR (hand_number_2 = ? AND hand_sute_2 = ?)
				  )
			)
		`, sessionID, card.Rank, card.Suit, card.Rank, card.Suit).Scan(&exists)
		if err != nil {
			return err
		}
		if exists {
			return errCardAlreadyUsed
		}
	}
	return nil
}
