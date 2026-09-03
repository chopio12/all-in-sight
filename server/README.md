# ポーカー記録システム（server）

## 必要なもの

- Python 3.10 以上（推奨: 3.12）
- pip

---

## セットアップ・起動手順

### 1. server ディレクトリへ移動

```bash
cd all-in-sight/server
```

すでに `server/` にいる場合はこの手順は不要。

### 2. 仮想環境を作成・有効化

```bash
python3.12 -m venv venv
source venv/bin/activate
```

プロンプトが `(venv)` で始まればOK。

### 3. ライブラリをインストール

```bash
pip install -r requirements.txt
```

### 4. サーバーを起動

```bash
uvicorn main:app --reload
```

起動に成功すると以下が表示される。

```
INFO:     Uvicorn running on http://127.0.0.1:8000 (Press CTRL+C to quit)
```

### 5. 動作確認

ブラウザで以下を開く。

```
http://localhost:8000/docs
```

Swagger UI が表示されれば起動成功。

---

## APIの使い方（Swagger UI）

1. `http://localhost:8000/docs` を開く
2. 試したいエンドポイントをクリック
3. `Try it out` をクリック
4. 値を入力して `Execute` をクリック

### POST /hands 入力例

```json
{
  "cards": [
    { "rank": "A", "suit": "s" },
    { "rank": "K", "suit": "h" }
  ]
}
```

---

## ファイル構成

```
server/
├── main.py
├── requirements.txt
├── poker.db
└── README.md
```

---

## サーバーの止め方

`Ctrl + C` で停止。仮想環境を抜ける場合は `deactivate` を実行。

---

## トラブルシューティング

### externally-managed-environment エラー

仮想環境外で `pip` を実行している可能性が高い。
`source venv/bin/activate` 実行後に再度インストールする。

### (venv) が表示されない

```bash
cd all-in-sight/server
source venv/bin/activate
```

### pydantic-core のビルドエラー

Python 3.12 を利用する。

```bash
brew install python@3.12
python3.12 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### ポート8000が使用中

```bash
python3 -m uvicorn main:app --reload --port 8001
```

### DBを初期化したい

```bash
rm poker.db
python3 -m uvicorn main:app --reload
```

再起動時に空のDBが自動作成される。
