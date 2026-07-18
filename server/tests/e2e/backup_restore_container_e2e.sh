#!/bin/sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
compose_file="$script_dir/backup/docker-compose.yml"
project_name="grtblog-backup-e2e-$$"
BACKUP_E2E_PORT=$((20000 + ($$ % 20000)))
export BACKUP_E2E_PORT
base_url="http://127.0.0.1:$BACKUP_E2E_PORT"
temp_dir=$(mktemp -d)

for command_name in docker curl jq tar; do
	if ! command -v "$command_name" >/dev/null 2>&1; then
		echo "[backup-e2e] missing required command: $command_name" >&2
		exit 1
	fi
done

compose() {
	docker compose -p "$project_name" -f "$compose_file" "$@"
}

cleanup() {
	exit_code=$?
	trap - EXIT INT TERM
	if [ "$exit_code" -ne 0 ]; then
		echo "[backup-e2e] failure; container logs follow" >&2
		compose logs --no-color >&2 || true
	fi
	compose down -v --remove-orphans >/dev/null 2>&1 || true
	rm -rf "$temp_dir"
	exit "$exit_code"
}
trap cleanup EXIT INT TERM

wait_for_api() {
	attempt=0
	while [ "$attempt" -lt 120 ]; do
		if curl -fsS "$base_url/health/liveness" >/dev/null 2>&1; then
			return 0
		fi
		attempt=$((attempt + 1))
		sleep 1
	done
	echo "[backup-e2e] API did not become live" >&2
	return 1
}

wait_for_backup() {
	backup_id=$1
	attempt=0
	while [ "$attempt" -lt 180 ]; do
		backup_json=$(curl -fsS -H "Authorization: Bearer $token" "$base_url/api/v2/admin/backups/$backup_id")
		backup_status=$(printf '%s' "$backup_json" | jq -r '.data.status')
		case "$backup_status" in
			completed) return 0 ;;
			failed)
				printf '%s\n' "$backup_json" >&2
				return 1
				;;
		esac
		attempt=$((attempt + 1))
		sleep 1
	done
	echo "[backup-e2e] backup $backup_id did not complete" >&2
	return 1
}

echo "[backup-e2e] building and starting isolated stack on port $BACKUP_E2E_PORT"
compose up --build -d
wait_for_api

curl -fsS -H 'Content-Type: application/json' \
	-d '{"username":"backupadmin","nickname":"Backup E2E","email":"backup-e2e@example.com","password":"BackupE2E!123"}' \
	"$base_url/api/v2/auth/register" >/dev/null
login_json=$(curl -fsS -H 'Content-Type: application/json' \
	-d '{"credential":"backupadmin","password":"BackupE2E!123"}' \
	"$base_url/api/v2/auth/login")
token=$(printf '%s' "$login_json" | jq -er '.data.token')

compose exec -T postgres psql -U postgres -d grtblog -v ON_ERROR_STOP=1 <<'SQL'
CREATE TABLE public.backup_e2e_probe (id INTEGER PRIMARY KEY, value TEXT NOT NULL);
INSERT INTO public.backup_e2e_probe (id, value) VALUES (1, 'before-backup');
SQL
compose exec -T -u app server sh -c 'mkdir -p /app/storage/uploads/e2e && printf %s before-backup-file > /app/storage/uploads/e2e/probe.txt'

create_json=$(curl -fsS -X POST -H "Authorization: Bearer $token" "$base_url/api/v2/admin/backups")
manual_backup_id=$(printf '%s' "$create_json" | jq -er '.data.id')
wait_for_backup "$manual_backup_id"

ticket_json=$(curl -fsS -X POST -H "Authorization: Bearer $token" \
	"$base_url/api/v2/admin/backups/$manual_backup_id/download-ticket")
download_path=$(printf '%s' "$ticket_json" | jq -er '.data.url')
curl -fsS "$base_url$download_path" -o "$temp_dir/site-backup.tar.gz"
tar -tzf "$temp_dir/site-backup.tar.gz" | grep -q '^database/site.dump$'
tar -tzf "$temp_dir/site-backup.tar.gz" | grep -q '^files/uploads/e2e/probe.txt$'
tar -xOzf "$temp_dir/site-backup.tar.gz" manifest.json | jq -e '.formatVersion == 1 and .uploadFileCount == 1' >/dev/null

curl -fsS -X PUT -H "Authorization: Bearer $token" -H 'Content-Type: application/json' \
	-d '{"enabled":true,"intervalHours":1,"retentionCount":2}' \
	"$base_url/api/v2/admin/backups/schedule" >/dev/null
compose exec -T postgres psql -U postgres -d grtblog -v ON_ERROR_STOP=1 \
	-c "UPDATE backup_ops.schedule_config SET next_run_at = NOW() - INTERVAL '1 second' WHERE id = 1" >/dev/null

scheduled_backup_id=""
attempt=0
while [ "$attempt" -lt 120 ]; do
	list_json=$(curl -fsS -H "Authorization: Bearer $token" "$base_url/api/v2/admin/backups")
	scheduled_backup_id=$(printf '%s' "$list_json" | jq -r '[.data[] | select(.triggerType == "scheduled")][0].id // empty')
	if [ -n "$scheduled_backup_id" ]; then
		break
	fi
	attempt=$((attempt + 1))
	sleep 1
done
if [ -z "$scheduled_backup_id" ]; then
	echo "[backup-e2e] scheduler did not create a backup" >&2
	exit 1
fi
wait_for_backup "$scheduled_backup_id"
curl -fsS -X PATCH -H "Authorization: Bearer $token" -H 'Content-Type: application/json' \
	-d '{"pinned":true}' "$base_url/api/v2/admin/backups/$scheduled_backup_id/pin" >/dev/null

compose exec -T postgres psql -U postgres -d grtblog -v ON_ERROR_STOP=1 \
	-c "UPDATE public.backup_e2e_probe SET value = 'after-backup' WHERE id = 1" >/dev/null
compose exec -T -u app server sh -c 'printf %s after-backup-file > /app/storage/uploads/e2e/probe.txt && printf %s extra-file > /app/storage/uploads/e2e/extra.txt'

server_id=$(compose ps -q server)
started_before=$(docker inspect -f '{{.State.StartedAt}}' "$server_id")
curl -fsS -X POST -H "Authorization: Bearer $token" -H 'Content-Type: application/json' \
	-d '{"confirmation":"OVERWRITE"}' \
	"$base_url/api/v2/admin/backups/$manual_backup_id/restore" >/dev/null

attempt=0
restore_state=""
while [ "$attempt" -lt 180 ]; do
	if restore_json=$(curl -fsS -H "Authorization: Bearer $token" "$base_url/api/v2/admin/backups/restore-status" 2>/dev/null); then
		restore_state=$(printf '%s' "$restore_json" | jq -r '.data.state')
		case "$restore_state" in
			succeeded) break ;;
			failed)
				printf '%s\n' "$restore_json" >&2
				exit 1
				;;
		esac
	fi
	attempt=$((attempt + 1))
	sleep 1
done
if [ "$restore_state" != "succeeded" ]; then
	echo "[backup-e2e] restore did not succeed" >&2
	exit 1
fi

started_after=$(docker inspect -f '{{.State.StartedAt}}' "$server_id")
if [ "$started_before" = "$started_after" ]; then
	echo "[backup-e2e] server container did not restart for offline restore" >&2
	exit 1
fi
database_value=$(compose exec -T postgres psql -U postgres -d grtblog -Atqc 'SELECT value FROM public.backup_e2e_probe WHERE id = 1')
if [ "$database_value" != "before-backup" ]; then
	echo "[backup-e2e] database was not restored: $database_value" >&2
	exit 1
fi
compose exec -T -u app server sh -c 'test "$(cat /app/storage/uploads/e2e/probe.txt)" = before-backup-file && test ! -e /app/storage/uploads/e2e/extra.txt'

echo "[backup-e2e] simulating a fresh install and restoring through the setup endpoint"
compose stop server >/dev/null
compose exec -T postgres psql -U postgres -d grtblog -v ON_ERROR_STOP=1 <<'SQL'
DROP SCHEMA public CASCADE;
DROP SCHEMA backup_ops CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO postgres;
GRANT ALL ON SCHEMA public TO public;
SQL
compose start server >/dev/null
wait_for_api
setup_json=$(curl -fsS "$base_url/api/v2/auth/setup-state")
if [ "$(printf '%s' "$setup_json" | jq -r '.data.hasUser')" != "false" ]; then
	echo "[backup-e2e] simulated fresh install unexpectedly has a user" >&2
	exit 1
fi
curl -fsS -F "archive=@$temp_dir/site-backup.tar.gz;type=application/gzip" \
	-F 'confirmation=OVERWRITE' "$base_url/api/v2/auth/setup-restore" >/dev/null

attempt=0
setup_restored="false"
while [ "$attempt" -lt 180 ]; do
	if setup_json=$(curl -fsS "$base_url/api/v2/auth/setup-state" 2>/dev/null); then
		setup_restored=$(printf '%s' "$setup_json" | jq -r '.data.hasAdmin')
		if [ "$setup_restored" = "true" ]; then
			break
		fi
	fi
	attempt=$((attempt + 1))
	sleep 1
done
if [ "$setup_restored" != "true" ]; then
	echo "[backup-e2e] setup restore did not recover the administrator" >&2
	exit 1
fi
database_value=$(compose exec -T postgres psql -U postgres -d grtblog -Atqc 'SELECT value FROM public.backup_e2e_probe WHERE id = 1')
if [ "$database_value" != "before-backup" ]; then
	echo "[backup-e2e] setup restore did not recover database content: $database_value" >&2
	exit 1
fi
compose exec -T -u app server sh -c 'test "$(cat /app/storage/uploads/e2e/probe.txt)" = before-backup-file'

echo "[backup-e2e] PASS: archive, download, scheduler, offline restore, uploads, and initial-setup restore"
