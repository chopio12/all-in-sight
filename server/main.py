from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field, field_validator
from typing import Literal
import sqlite3
import json
from datetime import datetime

app = FastAPI(title="ポーカー記録システム", version="0.2")

DB_PATH = "poker.db"

# ─────────────────────────────────────────
# DB初期化
# ─────────────────────────────────────────

def init_db():
    with sqlite3.connect(DB_PATH) as conn:
        conn.execute("""
            CREATE TABLE IF NOT EXISTS hands (
                id         INTEGER PRIMARY KEY AUTOINCREMENT,
                cards      TEXT    NOT NULL,
                created_at TEXT    NOT NULL DEFAULT (datetime('now'))
            )
        """)

init_db()

# ─────────────────────────────────────────
# スキーマ定義（バリデーション）
# ─────────────────────────────────────────

class Card(BaseModel):
    rank: Literal["2","3","4","5","6","7","8","9","T","J","Q","K","A"]
    suit: Literal["s","h","d","c"]

class HandRequest(BaseModel):
    cards: list[Card] = Field(
        default=[
            {"rank": "A", "suit": "s"},
            {"rank": "K", "suit": "h"},
        ]
    )

    @field_validator("cards")
    @classmethod
    def must_be_two_cards(cls, v):
        if len(v) != 2:
            raise ValueError("cards must contain exactly 2 cards")
        return v

class HandResponse(BaseModel):
    id: int
    cards: list[Card]
    created_at: str

# ─────────────────────────────────────────
# エンドポイント
# ─────────────────────────────────────────

@app.get("/health")
def health_check():
    """サーバーの起動確認"""
    return {"status": "ok"}


@app.post("/hands", response_model=HandResponse, status_code=201)
def create_hand(body: HandRequest):
    """ハンドを記録する（Androidから呼ぶ）"""
    cards_json = json.dumps([c.model_dump() for c in body.cards])
    created_at = datetime.utcnow().strftime("%Y-%m-%dT%H:%M:%SZ")

    with sqlite3.connect(DB_PATH) as conn:
        cursor = conn.execute(
            "INSERT INTO hands (cards, created_at) VALUES (?, ?)",
            (cards_json, created_at)
        )
        new_id = cursor.lastrowid

    return HandResponse(id=new_id, cards=body.cards, created_at=created_at)


@app.get("/hands", response_model=list[HandResponse])
def get_hands():
    """ハンド一覧を取得する（管理者画面から呼ぶ）"""
    with sqlite3.connect(DB_PATH) as conn:
        rows = conn.execute(
            "SELECT id, cards, created_at FROM hands ORDER BY id DESC"
        ).fetchall()

    result = []
    for row in rows:
        cards = [Card(**c) for c in json.loads(row[1])]
        result.append(HandResponse(id=row[0], cards=cards, created_at=row[2]))

    return result
