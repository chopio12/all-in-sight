package main

import (
	"fmt"
	// database/sql は、SQLite などのデータベースを操作する共通APIです。
	"database/sql"
	// encoding/json は、JSONとGoの構造体を相互に変換します。
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	// SQLiteドライバをdatabase/sqlへ登録します。直接は呼ばないため、名前を _ にします。
	_ "modernc.org/sqlite"
)

const (
	// Go専用SQLiteファイルへの、goディレクトリから見た相対パスです。
	dbPath = "../poker-go.db"
	// HTTPサーバーが待ち受けるポートです。Python版と同時に起動できるよう8001を使います。
	addr = ":8001"
)

// card は、1枚のトランプをJSONで表すための構造体です。
// jsonタグにより、GoのRankフィールドをJSONでは"rank"という名前にします。
type card struct {
	Rank string `json:"rank"`
	Suit string `json:"suit"`
}

// handRequest は、POST /hands のリクエスト本文の形です。
type handRequest struct {
	PlayerID string `json:"player_id"`
	Position string `json:"position"`
	Cards    []card `json:"cards"`
}

// handResponse は、作成・取得したハンドをクライアントに返すときの形です。
type handResponse struct {
	ID        int64  `json:"id"`
	SessionID int64  `json:"session_id"`
	PlayerID  string `json:"player_id"`
	Position  string `json:"position"`
	Cards     []card `json:"cards"`
	CreatedAt string `json:"created_at"`
}

// sessionResponse は、ポーカーセッションの状態をクライアントへ返す形です。
type sessionResponse struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// server は、各HTTPハンドラから使うデータベース接続をまとめます。
type server struct {
	db *sql.DB
}

func main() {
	fmt.Printf("Starting server on %s\n", addr)
	// SQLiteデータベースに接続します。ファイルがなければSQLiteが作成します。
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 初回起動時でも使えるよう、handsテーブルを作成します。
	if err := initDB(db); err != nil {
		log.Fatal(err)
	}

	// ハンドラがデータベースを使えるよう、server構造体を作ります。
	app := &server{db: db}
	// muxは、HTTPメソッドとURLパスを見て呼び出す関数を振り分けるルーターです。
	mux := http.NewServeMux()
	mux.HandleFunc("/health", app.healthCheck)
	mux.HandleFunc("/sessions/start", app.startSession)
	mux.HandleFunc("/sessions/end", app.endSession)
	mux.HandleFunc("/sessions/current", app.getCurrentSession)
	mux.HandleFunc("/hands", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			app.getHands(w, r)
		case http.MethodPost:
			app.createHand(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	// HTTPサーバーを起動します。この行はサーバーを停止するまで戻りません。
	log.Printf("Poker API listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func initDB(db *sql.DB) error {
	// セッションと手札を分け、手札は必ず1つのセッションに所属させます。
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

func (app *server) healthCheck(w http.ResponseWriter, _ *http.Request) {
	// サーバーが起動しているかを確認するため、固定のJSONを返します。
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

func (app *server) createHand(w http.ResponseWriter, r *http.Request) {
	// リクエスト本文は読み終えたら閉じます。
	defer r.Body.Close()

	// curl -d などで送られたJSONを、handRequest構造体に読み込みます。
	var request handRequest
	decoder := json.NewDecoder(r.Body)
	// API仕様にないフィールドがJSONに含まれていた場合はエラーにします。
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

	// カードが2枚か、rankとsuitが許可された値かを確認します。
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

	// UTC時刻をAPIで扱いやすいRFC3339形式（例: 2026-09-06T12:00:00Z）にします。
	createdAt := time.Now().UTC().Format(time.RFC3339)
	// ? は値を安全に渡すためのプレースホルダーです。SQL文へ値を直接つなげません。
	result, err := app.db.Exec(
		"INSERT INTO hands (session_id, player_id, position, hand_number_1, hand_number_2, hand_sute_1, hand_sute_2, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		sessionID, request.PlayerID, request.Position, request.Cards[0].Rank, request.Cards[1].Rank, request.Cards[0].Suit, request.Cards[1].Suit, createdAt,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save hand")
		return
	}

	// SQLiteが新しく割り当てたIDを取得します。
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
	// IDの大きい順、つまり新しく保存したハンドから取得します。
	query += "ORDER BY id DESC"
	rows, err := app.db.Query(query, queryArgs...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load hands")
		return
	}
	defer rows.Close()

	// 空の配列を先に作ることで、データが0件でもJSONでは [] を返します。
	hands := make([]handResponse, 0)
	for rows.Next() {
		var hand handResponse
		// SQLの各列をGoの変数へ読み取ります。& は変数に値を書き込む指定です。
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

func validateCards(cards []card) error {
	// ポーカーの手札は必ず2枚という、このAPIの入力ルールを確認します。
	if len(cards) != 2 {
		return errors.New("cards must contain exactly 2 cards")
	}
	for _, card := range cards {
		if !isValidRank(card.Rank) {
			return errors.New("rank must be one of 2-9, T, J, Q, K, A")
		}
		if !isValidSuit(card.Suit) {
			return errors.New("suit must be one of s, h, d, c")
		}
	}
	if cards[0].Rank == cards[1].Rank && cards[0].Suit == cards[1].Suit {
		return errors.New("cards must not contain the same card twice")
	}
	return nil
}

func isValidRank(rank string) bool {
	// 文字列に含まれているかで、使用できるランクかを判定します。
	return len(rank) == 1 && contains("23456789TJQKA", rank)
}

func isValidSuit(suit string) bool {
	// 文字列に含まれているかで、使用できるスートかを判定します。
	return len(suit) == 1 && contains("shdc", suit)
}

func contains(values, value string) bool {
	// valuesを1文字ずつ調べ、valueと一致する文字があればtrueを返します。
	for _, candidate := range values {
		if string(candidate) == value {
			return true
		}
	}
	return false
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	// レスポンスがJSONであることと、HTTPステータスコードを先に設定します。
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	// Goの値をJSONへ変換して、クライアントへ書き込みます。
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("failed to write JSON response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, detail string) {
	// エラーのJSON形式を {"detail":"..."} にそろえます。
	writeJSON(w, status, map[string]string{"detail": detail})
}
