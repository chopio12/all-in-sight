# ポーカー記録システム API 仕様書 v0.2

> **方針**: まず動くものを作る。データ項目・エンドポイントは後から追加していく。

---

## 基本情報

| 項目 | 値 |
|------|-----|
| Base URL (開発) | `http://localhost:8000` |
| Base URL (本番) | `https://your-app.railway.app` |
| フォーマット | JSON |
| 認証 | なし（v0.2は認証不要） |

---

## カードの表現形式

ハンドのカードは1枚ごとに `rank` と `suit` を持つオブジェクトで表現する。

### rank（数字・絵札）

| 値 | 意味 |
|----|------|
| `"2"` 〜 `"9"` | 数字札 |
| `"T"` | 10 |
| `"J"` | Jack |
| `"Q"` | Queen |
| `"K"` | King |
| `"A"` | Ace |

### suit（スート）

| 値 | 意味 |
|----|------|
| `"s"` | スペード ♠ |
| `"h"` | ハート ♥ |
| `"d"` | ダイヤ ♦ |
| `"c"` | クラブ ♣ |

### カードオブジェクト例

```json
{ "rank": "A", "suit": "s" }   // A♠
{ "rank": "K", "suit": "h" }   // K♥
{ "rank": "T", "suit": "d" }   // T♦
{ "rank": "2", "suit": "c" }   // 2♣
```

---

## エンドポイント一覧

### 1. ハンドを記録する

```
POST /hands
```

#### リクエスト

```json
{
  "cards": [
    { "rank": "A", "suit": "s" },
    { "rank": "K", "suit": "h" }
  ]
}
```

| フィールド | 型 | 必須 | 説明 |
|------------|-----|------|------|
| `cards` | array | ✅ | 手札のカード。必ず2枚 |
| `cards[].rank` | string | ✅ | `2`〜`9`, `T`, `J`, `Q`, `K`, `A` のいずれか |
| `cards[].suit` | string | ✅ | `s`, `h`, `d`, `c` のいずれか |

#### レスポンス（成功）

```json
HTTP 201 Created

{
  "id": 1,
  "cards": [
    { "rank": "A", "suit": "s" },
    { "rank": "K", "suit": "h" }
  ],
  "created_at": "2025-05-16T12:00:00Z"
}
```

#### レスポンス（バリデーションエラー）

```json
HTTP 422 Unprocessable Entity

{
  "detail": "cards must contain exactly 2 cards"
}
```

---

### 2. ハンド一覧を取得する

```
GET /hands
```

#### リクエスト

パラメータなし

#### レスポンス（成功）

```json
HTTP 200 OK

[
  {
    "id": 1,
    "cards": [
      { "rank": "A", "suit": "s" },
      { "rank": "K", "suit": "h" }
    ],
    "created_at": "2025-05-16T12:00:00Z"
  },
  {
    "id": 2,
    "cards": [
      { "rank": "7", "suit": "d" },
      { "rank": "2", "suit": "c" }
    ],
    "created_at": "2025-05-16T12:05:00Z"
  }
]
```

---

### 3. ヘルスチェック

```
GET /health
```

#### レスポンス

```json
HTTP 200 OK

{ "status": "ok" }
```

---

## DBスキーマ（SQLite）

カードはJSON文字列としてDBに保存する（シンプルさ優先）。

```sql
CREATE TABLE hands (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    cards      TEXT    NOT NULL,  -- '[{"rank":"A","suit":"s"},{"rank":"K","suit":"h"}]'
    created_at TEXT    NOT NULL DEFAULT (datetime('now'))
);
```

---

## 開発メモ

- Swagger UI（自動生成）: `http://localhost:8000/docs`
- ここでエンドポイントを直接試せる。メンバーAとBはまずここで動作確認すること。

---

## 今後追加予定（v0.2では不要）

- ボード（場のカード）
- BET額
- ポジション
- ポット額
- セッション / ハンド単位の管理
- 認証
