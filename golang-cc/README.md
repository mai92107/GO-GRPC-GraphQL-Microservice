# 信用卡最佳推薦 App

Gin、PostgreSQL 與 React 建置的信用卡推薦產品。管理員維護銀行、卡片、活動、消費類別與回饋單位；會員管理卡片夾、偏好、推薦與交易。

## 啟動本機 Demo

使用 Docker Compose 啟動可在 Docker Desktop 圖形介面查看的服務：

```bash
docker compose up
```

- Web App 與 API：<http://localhost:8080>
- Mailpit：<http://localhost:8029>
- PostgreSQL：`localhost:54329`

預設帳號：

- 管理員：`admin@example.test` / `admin-password-123`

PostgreSQL 僅在 `postgres-data` volume 第一次建立時執行 `docker/postgres-init/00-initialize.sh`。初始化流程會依序套用所有 `migrations/*.up.sql`，並建立兩個預設帳號；SQL 只保存 Argon2id 雜湊。後續容器重啟不會重新初始化或覆寫資料。

Server 啟動時只連線資料庫並啟動 HTTP server，不會執行 migration 或修改業務資料。未來 migration 必須透過獨立流程執行。

## 本機開發

先啟動 PostgreSQL 與 Mailpit，再執行 Go server：

```bash
docker compose up postgres mailpit
go run .
```

前端開發：

```bash
cd web
npm install
npm run dev
```

正式前端由 `npm run build` 輸出至 `web/dist`，並嵌入 Go server。

## 本地 CI/CD

本機可以用 Makefile 串起測試、編譯、Docker build 與 Docker Compose 部署：

```bash
make deploy
```

`make deploy` 會依序執行 Go 測試、前端測試、前端正式 build、Go build、`docker compose build app`，最後用 `docker compose up -d` 啟動或更新服務。預設不執行 `docker compose down`，避免每次部署都停止 PostgreSQL 與 Mailpit；需要完整停止時可另外執行：

```bash
make down
```

常用指令：

```bash
make test
make build
make docker-build
make status
make logs
```

若要在本機 Git push 前自動部署，可以安裝 Git hook：

```bash
./scripts/install-git-hooks.sh pre-push
```

也可以改成 commit 成功後部署：

```bash
./scripts/install-git-hooks.sh post-commit
```

Git hooks 存在於本機 `.git/hooks`，不會被 Git 自動同步到其他開發者電腦。安裝腳本可重複執行；若原本已有同名 hook，會先備份成 `.bak`。

## 設定

應用程式由 JSON 設定讀取 DB、Mail、Server、Monitoring 與 Telegram：

- 本機：`configs/local.json`
- Docker：`configs/docker.json`
- 測試：`configs/test.json`

PostgreSQL 設定拆分為 `username`、`password`、`host`、`port`、`database` 與 `sslmode`。`monitoring.token` 必須使用獨立監控 token，不得使用管理員密碼。

啟用 Telegram 推薦 Bot 時，設定 `telegram.enabled=true` 與 `telegram.bot_token`。管理員需先透過 `/api/admin/telegram-bindings` 將 Telegram `chat_id` 綁定至啟用中的會員。

## API 與權限

- Public：`/api/public/auth/...`
- Admin：`/api/admin/...`
- Member：`/api/member/...`
- Actuator：`/actuator/health`、`/actuator/prometheus`

Admin 不可呼叫 Member API，Member 不可呼叫 Admin API。Mutation API 必須帶登入或 `/api/public/auth/me` 回傳的 CSRF token：

```text
X-CSRF-Token: <token>
```

Actuator 僅接受 Admin session 或設定檔中的監控 Bearer token。Docker healthcheck 使用監控 token：

```bash
curl -H 'Authorization: Bearer local-monitoring-token-change-me' http://localhost:8080/actuator/health
curl -H 'Authorization: Bearer local-monitoring-token-change-me' http://localhost:8080/actuator/prometheus
```

## 後端架構

- `internal/server`：Gin engine、routes、middleware 與 Actuator。
- `internal/controllers/{public,admin,member}`：Gin 輸入驗證、Controller DTO 與 HTTP response。
- `internal/services/{public,admin,member}`：業務規則與流程，不操作 DB。
- `internal/repositories/{public,admin,member}`：唯一可操作 PostgreSQL 與 transaction 的業務層。
- `internal/domain`：跨層資料型別與業務錯誤，不含 HTTP JSON tag。
- `internal/utils`：Email sender、UUID、token 與 hash 工具。

卡片可建立多期不重疊活動，系統依交易日期選擇當期活動。活動內可建立多個優惠條件：

- 同一累計層只採回饋價值最高的條件，不同累計層可相加。
- 條件可設定個別月上限，活動亦可依回饋單位設定共用月上限。
- 活動必須同時符合消費類別與標準店家代碼；未設定店家代表所有店家適用。選擇「其他」時店家名稱僅供紀錄。
- 帳戶等級不符的優惠不計入推薦；需登錄、帳戶設定或切換 APP 方案的優惠會計入並顯示提醒。
- 未限定支付方式的規則會顯示為「不限支付方式」；只有特定支付方式產生不同回饋時才另外列出。選擇不限支付方式時，交易以 `any_payment` 保存。
- 推薦分數會先依管理者設定將回饋單位換算為台幣價值，再套用會員偏好權重。

## 驗證

```bash
go test -p 1 -count=1 -race ./...
go vet ./...
cd web && npm test -- --run && npm run build
docker compose config --quiet
docker compose build app
```
