# all-in-sight

ポーカーのハンドデータを記録・管理するシステム。

---

## 構成

```
all-in-sight/
├── server/         # バックエンド（FastAPI）
├── android/        # Androidアプリ（Kotlin）
├── web/            # 管理者Webアプリ（HTML + JS）
└── docs/           # ドキュメント
```

## 各コンポーネントの役割

| フォルダ | 役割 | 担当 |
|----------|------|------|
| `server/` | APIサーバー。ハンドデータの受信・保存・返却 | けいすけ |
| `android/` | ハンドデータをサーバーに送信するアプリ | ゆうま |
| `web/` | サーバーのデータを画面に表示する管理者アプリ | リーダーオブリーダー |

---

## ドキュメント

- [API仕様書](docs/poker_api_spec.md)

---

## 開発の始め方

各コンポーネントのセットアップ手順は、それぞれのフォルダ内の README を参照。

- [Go サーバーのセットアップ](server/go/README.md)
- [Python サーバーのセットアップ](server/python/README.md)
- Android のセットアップ（準備中）
- Web のセットアップ（準備中）

---

## 技術スタック

| コンポーネント | 技術 |
|----------------|------|
| サーバー | Go / SQLite、Python / FastAPI / SQLite |
| Androidアプリ | Kotlin(Java) / Jetpack Compose(どっちでも良い) |
| 管理者Webアプリ | HTML / JavaScript |
| インフラ | Railway |
# all-in-sight