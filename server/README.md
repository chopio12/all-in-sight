# ポーカー記録システム

## 必要なもの

- Python 3.10 以上
- pip3

---

## セットアップ・起動手順

### 1. リポジトリをクローン

```bash
git clone https://github.com/your-org/poker-system.git
cd poker-system
```

### 2. 仮想環境を作成・有効化

```bash
python3.12 -m venv venv
source venv/bin/activate
```

ターミナルのプロンプトが `(venv)` で始まればOKです。

```
(venv) $ 
```

> **注意**: 以降のコマンドは必ず `(venv)` が表示されている状態で実行すること。

### 3. ライブラリをインストール

```bash
pip install -r requirements.txt
```

### 4. サーバーを起動

```bash
uvicorn main:app --reload
```

起動に成功すると以下のように表示されます。

```
INFO:     Uvicorn running on http://127.0.0.1:8000 (Press CTRL+C to quit)
INFO:     Started reloader process
```

### 4. 動作確認

ブラウザで以下のURLを開く。

```
http://localhost:8000/docs
```

Swagger UI が表示されれば起動成功です。

---

## APIの使い方（Swagger UIで試す）

1. `http://localhost:8000/docs` を開く
2. 試したいエンドポイントをクリック
3. 「Try it out」ボタンをクリック
4. 必要な値を入力して「Execute」をクリック

### POST /hands（ハンドを記録する）の入力例

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
poker-system/
├── main.py           # サーバー本体
├── requirements.txt  # 必要なライブラリ
├── poker.db          # SQLiteデータベース（起動時に自動生成）
└── README.md
```

---

## サーバーの止め方

ターミナルで `Ctrl + C` を押す。仮想環境を抜ける場合はさらに `deactivate` を実行する。

---

## トラブルシューティング

### externally-managed-environment エラーが出る

仮想環境を使わずに直接 pip を実行しているのが原因。
セットアップ手順の通り `source venv/bin/activate` で仮想環境を有効化してから pip を実行する。

### (venv) が表示されない・仮想環境に入れていない

```bash
# プロジェクトのフォルダにいることを確認してから実行
source venv/bin/activate
```

### pydantic-core のビルドエラーが出る

Python 3.14 はまだ pydantic が未対応。Python 3.12 を使う。

```bash
brew install python@3.12
python3.12 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### ポート8000がすでに使われている

別のポートで起動する。

```bash
python3 -m uvicorn main:app --reload --port 8001
```

### poker.db を初期化したい（データを全部消したい）

```bash
rm poker.db
python3 -m uvicorn main:app --reload
```

再起動時に空のDBが自動で作成される。
