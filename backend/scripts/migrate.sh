#!/bin/sh
# Applies all *.up.sql migrations, in order, against the taskflow-postgres
# container started by `docker compose up -d` (run from Golang/Web/TaskFlow).
#
# Usage (from backend/): sh scripts/migrate.sh

set -e

for f in migrations/*.up.sql; do
  echo "Applying $f"
  cat "$f" | docker exec -i taskflow-postgres psql -U taskflow -d taskflow
done

echo "All migrations applied."
