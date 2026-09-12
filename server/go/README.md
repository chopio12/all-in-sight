# ポーカー API（Go）

ポーカー記録 API の Go 実装です。Go 専用の SQLite データベース `server/poker-go.db` を使用し、`http://localhost:8001` で待ち受けます。

## 必要なもの

- Go 1.27 以上

## 起動する

```bash
cd all-in-sight/server/go
go run .
```

初回起動時に `../poker-go.db` と必要なテーブルが作成されます。

サーバーの起動を確認します。

```bash
curl http://localhost:8001/health
```

```json
{"status":"ok"}
```

## セッションの流れ

サーバー全体で有効にできるポーカーセッションは1つだけです。クライアントは手札登録時に `session_id` を送る必要がありません。サーバーが有効なセッションへ手札を自動的に関連付けます。

手札を登録する前にセッションを開始します。

```bash
curl -i -X POST http://localhost:8001/sessions/start
```

```json
{"id":1,"status":"active","created_at":"2026-09-12T12:00:00Z"}
```

現在有効なセッションを確認します。

```bash
curl http://localhost:8001/sessions/current
```

ゲーム終了時にセッションを終了します。

```bash
curl -i -X POST http://localhost:8001/sessions/end
```

## 手札を登録する

```bash
curl -i -X POST http://localhost:8001/hands \
  -H 'Content-Type: application/json' \
  -d '{"player_id":"player-1","position":"BTN","cards":[{"rank":"A","suit":"s"},{"rank":"K","suit":"h"}]}'
```

成功すると `201 Created` が返ります。

```json
{"id":1,"session_id":1,"player_id":"player-1","position":"BTN","cards":[{"rank":"A","suit":"s"},{"rank":"K","suit":"h"}],"created_at":"2026-09-12T12:00:00Z"}
```

有効なセッションがない場合、`POST /hands` は `409 Conflict` を返します。

## 手札一覧を取得する

```bash
curl http://localhost:8001/hands
```

手札は ID の降順で返ります。

特定のセッションの手札だけを取得するには、`session_id` をクエリパラメータで指定します。

```bash
curl "http://localhost:8001/hands?session_id=1"
```

`session_id` を省略すると、すべてのセッションの手札を返します。`session_id` は1以上の整数で指定し、不正な値の場合は `422 Unprocessable Entity` を返します。

## 入力ルール

- `player_id` は必須です。
- `position` は必須です。例: `BTN`、`SB`、`BB`。
- `cards` は必ず2枚です。
- カードの `rank` は `2` から `9`、`T`、`J`、`Q`、`K`、`A` のいずれかです。
- カードの `suit` は `s`、`h`、`d`、`c` のいずれかです。
- 1つの手札内で、同じ `rank` と `suit` のカードを2枚送ることはできません。重複時は `422 Unprocessable Entity` を返します。
- 同じセッションでは、同じ `position` の手札を複数登録できません。
- 同じセッションでは、同じ `rank` と `suit` のカードを複数の手札へ登録できません。重複時は `409 Conflict` を返します。

## テストを実行する

```bash
cd all-in-sight/server/go
go test -v ./...
```

## データベースを初期化する

Go サーバーを停止してから実行します。

```bash
cd all-in-sight/server
rm poker-go.db
```

次回起動時に空のデータベースが作成されます。
