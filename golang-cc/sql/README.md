# SQL Artifacts

這裡整理兩種資料庫 SQL 版本：

- Migration 版：使用 `../migrations/*.up.sql` 與 `../migrations/*.down.sql`，依檔名前綴版本號逐版升級既有資料庫。
- Initial 版：使用 `initial.sql`，可在全新的 PostgreSQL database 一次建立到目前最新版 schema 與種子資料。

`initial.sql` 是將 `migrations/000001` 到 `migrations/000025` 套用到空資料庫後，用 `pg_dump --no-owner --no-privileges --format=plain` 產出的 snapshot。檔案內包含 `schema_migrations` 版本紀錄，因此用它初始化的新資料庫會被視為已套用到目前 migration 版本。

本機重新產生方式：

```bash
docker compose up -d --wait postgres
docker compose exec -T postgres sh -c 'dropdb -U credit_cards --if-exists credit_cards_initial_snapshot && createdb -U credit_cards credit_cards_initial_snapshot'
docker compose exec -T postgres sh -c 'set -eu; for migration in /migrations/*.up.sql; do version=$(basename "$migration" | cut -d_ -f1); psql --set ON_ERROR_STOP=on -U credit_cards -d credit_cards_initial_snapshot -f "$migration"; psql --set ON_ERROR_STOP=on -U credit_cards -d credit_cards_initial_snapshot -c "INSERT INTO schema_migrations(version) VALUES ($version) ON CONFLICT DO NOTHING"; done'
docker compose exec -T postgres pg_dump -U credit_cards -d credit_cards_initial_snapshot --no-owner --no-privileges --format=plain > sql/initial.sql
```
