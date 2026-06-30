## AI 開發規範

本專案允許使用 AI 協助開發，但任何產出必須符合既有架構與邊界，不得為了快速完成而破壞分層、命名、交易一致性或安全設計。

### 分層規則

AI 修改程式時必須遵守以下責任邊界：

- Controller 只處理：
  - Gin request / response
  - path、query、body 綁定
  - HTTP 狀態碼
  - DTO 轉換
  - 呼叫 Service

- Service 只處理：
  - 業務流程
  - 權限與狀態檢查
  - 跨 Repository 的流程編排
  - domain error 回傳

- Repository 只處理：
  - PostgreSQL 查詢
  - transaction
  - row scan
  - SQL error 轉換

- Domain 只放：
  - 業務資料結構
  - enum / const
  - 業務錯誤
  - 不得包含 HTTP JSON tag
  - 不得引用 Gin、pgx、gorm、React 相關套件

### 禁止事項

AI 不得執行以下行為：

- 不得在 Controller 直接查 DB。
- 不得在 Service 直接寫 SQL。
- 不得在 Repository 寫 HTTP response。
- 不得在 Server 啟動時自動執行 migration。
- 不得在 migration 以外的地方修改 schema。
- 不得將 Admin API 與 Member API 混用。
- 不得讓 Admin session 呼叫 Member API。
- 不得讓 Member session 呼叫 Admin API。
- 不得略過 CSRF 檢查。
- 不得將管理員密碼當作 monitoring token。
- 不得儲存信用卡完整卡號、CVV、有效期限。
- 不得將密碼、token、secret 寫死在程式碼。
- 不得為了解決型別錯誤而任意把 uuid 改成 text。
- 不得在 SQL 中混用 bigint id 與 uuid id。
- 不得用 `SELECT *`。
- 不得忽略 transaction rollback / commit error。
- 不得吞掉 error 或只用 `println` 處理錯誤。

### Go 後端規範

- 所有 public function 必須回傳明確 error。
- context 必須由上層傳入，不得在業務邏輯中任意使用 `context.Background()`。
- 金額、回饋率、點數價值不得使用 float，必須使用 decimal 或字串轉換後處理。
- 時間判斷必須使用傳入的 `at time.Time`，不得在推薦邏輯中直接使用 `time.Now()`。
- Repository 查詢必須明確列出欄位。
- SQL 查詢需要避免 N+1。
- 多筆寫入必須使用 transaction。
- 可預期的業務錯誤應轉成 domain error，不應直接暴露 PostgreSQL error。

### React 前端規範

- 不得把多個頁面用大量 if/else 寫在同一個 component。
- 頁面應拆分為：
  - page component
  - form component
  - table/list component
  - API client
  - type definition
- API response type 應集中管理，不得在各 component 重複定義。
- 表單驗證邏輯不得散落在 JSX 中。
- loading、error、empty state 必須明確處理。
- 不得直接在 component 中硬寫 API URL。

### 信用卡推薦邏輯規範

推薦邏輯必須維持以下模型：

- Card：信用卡本體。
- Reward Plan：特定期間有效的回饋方案。
- Reward Component：方案中的單一回饋條件。
- Requirement：回饋條件需要符合的資格。
- Layer：可累加的回饋層。
- Stack Group：同一組內只取最佳回饋。
- Cap：回饋上限。
- Reminder：需要登錄、切換 APP、帳戶等級等提醒。

計算規則：

- 同一 stack group 只取價值最高的 reward component。
- 不同 layer 可累加。
- 帳戶等級不符者不得計入推薦。
- 支付方式不符者不得計入推薦。
- 需登錄或需設定者可以計入，但必須顯示提醒。
- 未限制支付方式時，以 `any_payment` 表示。
- 回饋單位必須先依管理者設定換算成台幣價值，再套用會員偏好權重。
- 不得只回傳最終金額，必須保留可解釋的 reward component 明細。

### Migration 規範

- migration 檔案只能新增，不得修改已套用的舊 migration。
- 每個 migration 必須同時提供 `.up.sql` 與 `.down.sql`。
- schema 修改必須明確處理既有資料。
- 不得在 application 啟動時自動執行 migration。
- seed data 與 migration 應分離。
- 預設帳號密碼只能保存 Argon2id hash。

### AI 修改前檢查

AI 在修改程式前，必須先確認：

- 這次修改屬於 Controller、Service、Repository、Domain、Frontend 哪一層。
- 是否需要 migration。
- 是否會影響既有 API response。
- 是否會影響推薦計算邏輯。
- 是否需要補測試。
- 是否會破壞 Admin / Member 權限隔離。

### AI 回覆格式

AI 回覆程式修改建議時，請固定包含：

1. 修改目標
2. 影響範圍
3. 檔案清單
4. 修改後程式碼
5. 測試方式
6. 可能風險
7. 是否需要 migration

範例：

```text
請依照本專案 AI 開發規範修改。
本次目標是：新增會員卡片回饋快取查詢 API。
請先說明影響範圍，再提供完整程式碼。
不得跨層操作 DB，不得修改既有 migration，不得改變既有 API response 結構。