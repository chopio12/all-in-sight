# ポーカー API（Python）

ポーカー記録 API の FastAPI 実装です。Python 専用の SQLite データベース `server/poker.db` を使用し、`http://localhost:8000` で待ち受けます。

## 必要なもの

- Python 3.10 以上
- pip

## 初回セットアップと起動

```bash
cd all-in-sight/server/python
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
uvicorn main:app --reload --port 8000
```

初回起動時に `../poker.db` と `hands` テーブルが作成されます。

ブラウザで `http://localhost:8000/docs` を開くと、Swagger UI から API を試せます。

サーバーの起動を確認します。

```bash
curl http://localhost:8000/health
```

```json
{"status":"ok"}
```

## 手札を登録する

```bash
curl -i -X POST http://localhost:8000/hands \
  -H 'Content-Type: application/json' \
  -d '{"cards":[{"rank":"A","suit":"s"},{"rank":"K","suit":"h"}]}'
```

成功すると `201 Created` が返ります。

```json
{"id":1,"cards":[{"rank":"A","suit":"s"},{"rank":"K","suit":"h"}],"created_at":"2026-09-12T12:00:00Z"}
```

## 手札一覧を取得する

```bash
curl http://localhost:8000/hands
```

手札は ID の降順で返ります。

## 入力ルール

- `cards` は必ず2枚です。
- カードの `rank` は `2` から `9`、`T`、`J`、`Q`、`K`、`A` のいずれかです。
- カードの `suit` は `s`、`h`、`d`、`c` のいずれかです。

## データベースを初期化する

Python サーバーを停止してから実行します。

```bash
cd all-in-sight/server
rm poker.db
```

次回起動時に空のデータベースが作成されます。
