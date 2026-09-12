package main

import "database/sql"

type card struct {
	Rank string `json:"rank"`
	Suit string `json:"suit"`
}

type handRequest struct {
	PlayerID string `json:"player_id"`
	Position string `json:"position"`
	Cards    []card `json:"cards"`
}

type handResponse struct {
	ID        int64  `json:"id"`
	SessionID int64  `json:"session_id"`
	PlayerID  string `json:"player_id"`
	Position  string `json:"position"`
	Cards     []card `json:"cards"`
	CreatedAt string `json:"created_at"`
}

type sessionResponse struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type server struct {
	db *sql.DB
}
