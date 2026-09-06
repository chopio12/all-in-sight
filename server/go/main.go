package main

import (
	// database/sql は、SQLite などのデータベースを操作する共通APIです。
	"database/sql"
	// encoding/json は、JSONとGoの構造体を相互に変換します。
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	// SQLiteドライバをdatabase/sqlへ登録します。直接は呼ばないため、名前を _ にします。
	_ "modernc.org/sqlite"
)

const (
	// GoとPythonが共通で使うSQLiteファイルへの、goディレクトリから見た相対パスです。
	dbPath = "../poker.db"
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
	Cards []card `json:"cards"`
}

// handResponse は、作成・取得したハンドをクライアントに返すときの形です。
type handResponse struct {
	ID        int64  `json:"id"`
	Cards     []card `json:"cards"`
	CreatedAt string `json:"created_at"`
}

// server は、各HTTPハンドラから使うデータベース接続をまとめます。
type server struct {
	db *sql.DB
}

func main() {
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
	mux.HandleFunc("GET /health", app.healthCheck)
	mux.HandleFunc("GET /hands", app.getHands)
	mux.HandleFunc("POST /hands", app.createHand)

	// HTTPサーバーを起動します。この行はサーバーを停止するまで戻りません。
	log.Printf("Poker API listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func initDB(db *sql.DB) error {
	// IF NOT EXISTS があるため、すでにテーブルがあってもエラーにはなりません。
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS hands (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			cards      TEXT    NOT NULL,
			created_at TEXT    NOT NULL DEFAULT (datetime('now'))
		)
	`)
	return err
}

func (app *server) healthCheck(w http.ResponseWriter, _ *http.Request) {
	// サーバーが起動しているかを確認するため、固定のJSONを返します。
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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

	// カードが2枚か、rankとsuitが許可された値かを確認します。
	if err := validateCards(request.Cards); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	// SQLiteにはカードの配列をJSON文字列として保存します。
	cardsJSON, err := json.Marshal(request.Cards)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to encode cards")
		return
	}

	// UTC時刻をAPIで扱いやすいRFC3339形式（例: 2026-09-06T12:00:00Z）にします。
	createdAt := time.Now().UTC().Format(time.RFC3339)
	// ? は値を安全に渡すためのプレースホルダーです。SQL文へ値を直接つなげません。
	result, err := app.db.Exec(
		"INSERT INTO hands (cards, created_at) VALUES (?, ?)",
		string(cardsJSON),
		createdAt,
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
		Cards:     request.Cards,
		CreatedAt: createdAt,
	})
}

func (app *server) getHands(w http.ResponseWriter, _ *http.Request) {
	// IDの大きい順、つまり新しく保存したハンドから取得します。
	rows, err := app.db.Query("SELECT id, cards, created_at FROM hands ORDER BY id DESC")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load hands")
		return
	}
	defer rows.Close()

	// 空の配列を先に作ることで、データが0件でもJSONでは [] を返します。
	hands := make([]handResponse, 0)
	for rows.Next() {
		var hand handResponse
		var cardsJSON string
		// SQLの各列をGoの変数へ読み取ります。& は変数に値を書き込む指定です。
		if err := rows.Scan(&hand.ID, &cardsJSON, &hand.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to read hand")
			return
		}
		// DB内のJSON文字列を、レスポンス用のカード配列へ戻します。
		if err := json.Unmarshal([]byte(cardsJSON), &hand.Cards); err != nil {
			writeError(w, http.StatusInternalServerError, "stored hand contains invalid cards")
			return
		}
		hands = append(hands, hand)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load hands")
		return
	}

	writeJSON(w, http.StatusOK, hands)
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
