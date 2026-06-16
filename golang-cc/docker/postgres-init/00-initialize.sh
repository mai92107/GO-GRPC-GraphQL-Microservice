#!/bin/sh
set -eu

for migration in /migrations/*.up.sql; do
  version="$(basename "$migration" | cut -d_ -f1)"
  psql --set ON_ERROR_STOP=on --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" --file "$migration"
  psql --set ON_ERROR_STOP=on --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
    --command "INSERT INTO schema_migrations(version) VALUES ($version) ON CONFLICT DO NOTHING"
done

psql --set ON_ERROR_STOP=on --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<'SQL'
INSERT INTO users(id, email, password_hash, display_name, role, status) VALUES
  (
    '00000000-0000-0000-0000-000000000001',
    'admin@example.test',
    '$argon2id$v=19$m=65536,t=1,p=4$Z+G3h0nqVnd1xm4coM7o1w$zTIwi5D2u7S4Twfau1HgZZ43n024DBYgHq+KvXb5FuU',
    'Admin',
    'admin',
    'active'
  ),
  (
    '00000000-0000-0000-0000-000000000002',
    'member@example.test',
    '$argon2id$v=19$m=65536,t=1,p=4$PYZ76JtRapie+SCPFSxf0Q$/RYtLHz69J8P6bfPuSAV+CfGLlZXVs9caj1ca4IIfWw',
    'Member',
    'member',
    'active'
  )
ON CONFLICT (email) DO NOTHING;
SQL
