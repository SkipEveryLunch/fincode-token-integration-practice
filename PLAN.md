# fincode トークン決済（3Dセキュア対応）実装計画

## 概要

教育目的のプロジェクト。fincodeのカード登録＋トークン決済（3DS2.0対応）の一連のフローを学ぶ。

- **フロント**: React (Vite) / ローカル起動 (localhost:5173)
- **サーバー**: Go + Gin + PostgreSQL / Docker Compose (localhost:8080)
- **3DSコールバック**: ngrokでバックエンドを一時公開

---

## ディレクトリ構成

```
fincode-token-practice/
├── frontend/                  # React (Vite)
│   ├── index.html
│   ├── package.json           # @fincode/js を含む
│   └── src/
│       ├── main.tsx
│       ├── App.tsx
│       ├── api/
│       │   └── client.ts      # バックエンドAPI呼び出し関数
│       └── pages/
│           ├── CardRegisterPage.tsx   # カード登録フォーム
│           ├── CardConfirmPage.tsx    # カード確認画面
│           └── PurchasePage.tsx       # 「寿司 500円」購入画面
├── server/                    # Go + Gin
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── main.go
│   ├── handler/
│   │   ├── card.go            # POST /api/cards, GET /api/cards/active
│   │   └── payment.go         # POST /api/payments, GET /api/tds/callback
│   ├── service/
│   │   └── fincode.go         # fincodeへのHTTPリクエスト関数群
│   ├── repository/
│   │   ├── customer.go
│   │   ├── card.go
│   │   └── payment.go
│   ├── model/
│   │   ├── customer.go
│   │   ├── card.go
│   │   └── payment.go
│   └── db/
│       └── db.go              # DB接続 + マイグレーション
├── migrations/
│   └── 001_init.sql
├── docker-compose.yml
├── .env.example
└── PLAN.md
```

---

## DBスキーマ

```sql
-- シングルトンレコード。常に0〜1行。
-- fincode上のcustomer_idを保持するだけ。
CREATE TABLE customers (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    fincode_customer_id VARCHAR(64) NOT NULL UNIQUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- fincodeに登録したカード情報（マスク済み）。
-- is_alive=true のレコードが常に1件だけ（有効カード）。
-- カード更新時は旧レコードをis_alive=falseにしてから新規INSERT。
CREATE TABLE cards (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id         UUID NOT NULL REFERENCES customers(id),
    fincode_card_id     VARCHAR(64) NOT NULL,
    card_no_mask        VARCHAR(32),   -- 例: 456789******1234
    expire              VARCHAR(8),    -- 例: 2512
    brand               VARCHAR(16),   -- 例: VISA
    is_alive            BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 購入記録
CREATE TABLE payments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id             UUID NOT NULL REFERENCES cards(id),
    fincode_payment_id  VARCHAR(64) NOT NULL,
    fincode_access_id   VARCHAR(64) NOT NULL,
    amount              INTEGER NOT NULL DEFAULT 500,
    status              VARCHAR(32) NOT NULL DEFAULT 'UNPROCESSED',
    -- UNPROCESSED / CAPTURED / FAILED
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## バックエンドAPIエンドポイント

| Method | Path | 説明 |
|--------|------|------|
| `POST` | `/api/cards` | カスタマー取得or作成＋カード登録（upsert）を一括処理 |
| `GET`  | `/api/cards/active` | 有効カード（is_alive=true）1件取得 |
| `POST` | `/api/payments` | 決済登録 + 3DS認証開始 → redirect_urlを返す |
| `GET`  | `/api/tds/callback` | 3DS後のリダイレクト受け取り → 決済確定 |

### `POST /api/cards` の内部処理

```
リクエスト: { token }   ← フロントは token だけ渡す

バックエンド処理:
  1. customersテーブルを SELECT
       └ 0行 → fincode POST /v1/customers { email: "test@example.com" }
                → fincode_customer_id を取得して INSERT
       └ 1行 → そのまま fincode_customer_id を再利用

  2. fincode POST /v1/customers/{fincode_customer_id}/cards { token, ... }
       → fincode_card_id, card_no_mask, expire, brand を取得

  3. DBトランザクション:
       UPDATE cards SET is_alive = false WHERE is_alive = true
       INSERT INTO cards (customer_id, fincode_card_id, ..., is_alive = true)

レスポンス: { card_no_mask, expire, brand }
```

### `POST /api/payments` の内部処理

```
リクエスト: {} （ボディなし）

バックエンド処理:
  1. customers テーブルから fincode_customer_id を取得（なければ 400）
  2. cards テーブルから is_alive=true のカードを取得（なければ 400）
  3. fincode POST /v1/payments → { fincode_payment_id, fincode_access_id }
  4. fincode PUT /v1/payments/{id}/3dsecure2 { return_url: ngrok/.../callback }
       → { redirect_url }
  5. payments テーブルに INSERT（status=UNPROCESSED）
  6. レスポンス: { redirect_url }
```

---

## 画面フローと実装詳細

### 1. カード登録フォーム（CardRegisterPage）

```
[ユーザー操作]
カード番号・有効期限・セキュリティコード・名義人を入力 → 「確認」ボタン
```

**実装ポイント:**
- `@fincode/js` の UIコンポーネントをマウント → ユーザーがカード情報を入力
- 「確認」押下時に `getCardToken()` でカード情報をトークン化
- トークンはfincodeサーバーへ直接送信されるため、自社サーバーにカード番号は届かない
- `token` と表示用の `card_no_mask` / `expire` / `brand` を Reactの state に保持して確認画面へ遷移

**`@fincode/js` の使い方:**
```typescript
import { initFincode, getCardToken } from "@fincode/js";

const fincode = await initFincode({
  publicKey: import.meta.env.VITE_FINCODE_PUBLIC_KEY,
  isLiveMode: false,
});

const ui = fincode.ui({ layout: "vertical" });
ui.create("payment", { layout: "vertical" });
ui.mount("fincode-form-container", "400");

// 「確認」ボタン押下時
const res = await getCardToken(fincode, ui, "1");
const { token, card_no_mask, expire, brand } = res.list[0];
```

---

### 2. カード確認画面（CardConfirmPage）

```
[表示内容]
カード番号（マスク）: 411111******1111
有効期限: 25/12
ブランド: VISA

[「登録する」ボタン]
```

**「登録する」押下時:**
- `POST /api/cards` に `{ token }` を送信
- 成功したら購入画面へ遷移

---

### 3. 購入画面（PurchasePage）

```
[表示内容]
商品: 寿司
金額: ¥500

登録済みカード: VISA ****1111  ← GET /api/cards/active で取得

[「購入する」ボタン]
```

**「購入する」押下時のフロー:**

```
① フロント → バックエンド POST /api/payments （ボディなし）
    └ バックエンド:
        a. DB から active card 取得
        b. fincode POST /v1/payments           → { payment_id, access_id }
        c. fincode PUT  /v1/payments/{id}/3ds2  → { redirect_url }
           ※ return_url = ngrok経由の /api/tds/callback
        d. DB に payment INSERT (UNPROCESSED)
    └ レスポンス: { redirect_url }

② フロント: window.location.href = redirect_url
    └ ブラウザが fincode の 3DS 認証ページへ遷移

③-a frictionless の場合
    └ fincode が自動認証 → /api/tds/callback へリダイレクト

③-b challenge の場合
    └ ユーザーがカード会社の認証画面で認証
    └ 認証完了 → /api/tds/callback へリダイレクト

④ バックエンド GET /api/tds/callback（ngrok経由）
    ├ クエリパラメータから payment_id（fincode_payment_id）を特定
    ├ DB から該当 payment と card の fincode_card_id, customer の fincode_customer_id を取得
    ├ fincode PUT /v1/payments/{id} で決済確定
    │   { pay_type, access_id, customer_id, card_id }
    ├ DB の payment status を CAPTURED / FAILED に更新
    └ フロントの完了ページへリダイレクト (localhost:5173/complete?status=...)
```

---

## fincodeへのAPIリクエスト詳細

### 共通ヘッダー

```
Authorization: Bearer {シークレットキー}
Content-Type: application/json
```

### 1. カスタマー作成

```
POST https://api.test.fincode.jp/v1/customers
Body: { "email": "test@example.com" }

Response: { "id": "c_****" }
```

### 2. カード登録

```
POST https://api.test.fincode.jp/v1/customers/{fincode_customer_id}/cards
Body:
{
  "default_flag": "1",
  "token": "{getCardToken() で取得したトークン}"
}

Response:
{
  "id": "cs_****",
  "card_no_mask": "411111******1111",
  "expire": "2512",
  "brand": "Visa"
}
```

> ⚠️ `expire` / `holder_name` / `security_code` がトークンに含まれるか別途必要かは要確認。

### 3. 決済登録

```
POST https://api.test.fincode.jp/v1/payments
Body:
{
  "pay_type": "Card",
  "job_code": "CAPTURE",
  "amount": "500",
  "tds_type": "2",
  "tds2_type": "2"
}

Response: { "id": "p_****", "access_id": "ac_****" }
```

### 4. 3DS認証開始

```
PUT https://api.test.fincode.jp/v1/payments/{id}/3dsecure2
Body:
{
  "pay_type": "Card",
  "access_id": "ac_****",
  "return_url": "https://{ngrok_url}/api/tds/callback",
  "return_url_on_failure": "https://{ngrok_url}/api/tds/callback"
}

Response: { "redirect_url": "https://..." }
```

> ⚠️ エンドポイントのパスは公式リファレンスで要確認。
> Node.js SDK では `payments.execute3DSecureAuth(access_id, { return_url })` に対応。

### 5. 3DS後の決済確定

```
PUT https://api.test.fincode.jp/v1/payments/{id}
Body:
{
  "pay_type": "Card",
  "access_id": "ac_****",
  "customer_id": "c_****",
  "card_id": "cs_****"
}

Response: { "status": "CAPTURED", "amount": "500" }
```

> Node.js SDK では `payments.executeAfter3DSecureAuth(payment_id, { ... })` に対応。

---

## Docker Compose構成

```yaml
services:
  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: fincode_practice
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      retries: 5

  server:
    build: ./server
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://postgres:postgres@db:5432/fincode_practice?sslmode=disable
      FINCODE_SECRET_KEY: ${FINCODE_SECRET_KEY}
      FINCODE_API_BASE: https://api.test.fincode.jp
      TDS_RETURN_URL: ${TDS_RETURN_URL}
      FRONTEND_URL: http://localhost:5173
    depends_on:
      db:
        condition: service_healthy

volumes:
  pgdata:
```

---

## 環境変数

### server/.env
```
FINCODE_SECRET_KEY=m_test_xxxxxxxxxxxx
TDS_RETURN_URL=https://xxxx.ngrok-free.app/api/tds/callback
```

### frontend/.env.local
```
VITE_FINCODE_PUBLIC_KEY=p_test_xxxxxxxxxxxx
VITE_API_BASE_URL=http://localhost:8080
```

---

## 起動手順

```bash
# 1. ngrokでバックエンドを公開（先に起動しておく）
ngrok http 8080
# → 表示されたURLを server/.env の TDS_RETURN_URL に設定

# 2. バックエンド + DB
docker compose up --build

# 3. フロント（別ターミナル）
cd frontend
npm install
npm run dev
```

---

## 実装ステップ（推奨順）

| # | 対象 | 内容 |
|---|------|------|
| 1 | インフラ | `docker-compose.yml` + `migrations/001_init.sql` |
| 2 | サーバー骨格 | `main.go`、Gin ルーティング、DB接続 |
| 3 | fincode service層 | `service/fincode.go`（HTTPクライアント共通処理） |
| 4 | カード登録API | `POST /api/cards`（customer upsert + card upsert） |
| 5 | 有効カード取得API | `GET /api/cards/active` |
| 6 | 決済登録API | `POST /api/payments`（3DS開始まで含む） |
| 7 | 3DSコールバック | `GET /api/tds/callback`（決済確定） |
| 8 | フロント骨格 | Vite + React Router、3画面のルーティング |
| 9 | カード登録フォーム | `@fincode/js` 組み込み、`getCardToken()` |
| 10 | カード確認・登録 | `POST /api/cards` 呼び出し |
| 11 | 購入画面 | `POST /api/payments` → redirect_urlへ遷移 |
| 12 | 完了画面 | `/complete` でステータス表示 |
| 13 | 動作確認 | fincodeテスト用カード番号で一連フロー確認 |

---

## 要確認事項（実装時に公式ドキュメントで確認）

| # | 項目 | 内容 |
|---|------|------|
| 1 | 3DS認証開始エンドポイント | パスと正確なリクエストボディ |
| 2 | 決済確定エンドポイント | 3DS後の PUT ボディパラメータ |
| 3 | コールバックのHTTPメソッド | GET か POST か |
| 4 | カード登録ボディ | token にどこまで含まれるか（expire等） |
| 5 | テスト用カード番号 | 3DS challenge / frictionless の切り替え方法 |

---

## 参考リンク

- [fincode API Reference](https://docs.fincode.jp/api)
- [fincode-sdk-node (GitHub)](https://github.com/fincode-byGMO/fincode-sdk-node)
- [fincode-sdk-js (GitHub)](https://github.com/fincode-byGMO/fincode-sdk-js)
- [GoとReactで実装するサンプル (Zenn)](https://zenn.dev/fincode/articles/fincode-card-payment-cit)
- [Qiita: 処理の流れまとめ](https://qiita.com/Tatsu24/items/e0ae44bca65b05198af7)
