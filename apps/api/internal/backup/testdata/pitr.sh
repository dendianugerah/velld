#!/bin/sh
# End-to-end point-in-time recovery exercise, run inside a postgres container.
#
# Flow:
#   1. initdb + start a primary with archive_mode = on, archiving into ./archive
#   2. take a base backup with pg_basebackup
#   3. insert a "before" row (must survive recovery)
#   4. record a target timestamp
#   5. insert an "after" row (must NOT survive recovery)
#   6. force a WAL switch so the segment is archived
#   7. stop, restore the base backup into a fresh data dir, point it at the
#      archive, and recover to the target timestamp
#   8. assert the before row exists and the after row does not
set -eu

PGDATA_PRIMARY=/tmp/primary
PGDATA_RESTORE=/tmp/restore
ARCHIVE=/tmp/archive
SOCKET=/tmp/sock
PORT=5599

rm -rf "$PGDATA_PRIMARY" "$PGDATA_RESTORE" "$ARCHIVE" "$SOCKET"
mkdir -p "$ARCHIVE" "$SOCKET"

initdb -D "$PGDATA_PRIMARY" -U postgres >/dev/null

cat >> "$PGDATA_PRIMARY/postgresql.conf" <<EOF
port = $PORT
unix_socket_directories = '$SOCKET'
wal_level = replica
archive_mode = on
archive_command = 'test ! -f $ARCHIVE/%f && cp %p $ARCHIVE/%f'
max_wal_senders = 4
listen_addresses = ''
EOF

pg_ctl -D "$PGDATA_PRIMARY" -w start >/dev/null

psql() {
  command psql -h "$SOCKET" -p "$PORT" -U postgres -v ON_ERROR_STOP=1 "$@"
}

psql -d postgres -c "CREATE TABLE t (id int primary key, note text)" >/dev/null
psql -d postgres -c "INSERT INTO t VALUES (1, 'before')" >/dev/null

pg_basebackup -h "$SOCKET" -p "$PORT" -U postgres -D "$PGDATA_RESTORE" -X stream >/dev/null

TARGET_TIME=$(psql -d postgres -t -A -c "SELECT now()")
sleep 1
psql -d postgres -c "INSERT INTO t VALUES (2, 'after')" >/dev/null

psql -d postgres -c "SELECT pg_switch_wal()" >/dev/null
sleep 1

pg_ctl -D "$PGDATA_PRIMARY" -m fast -w stop >/dev/null

cat >> "$PGDATA_RESTORE/postgresql.auto.conf" <<EOF
restore_command = 'cp $ARCHIVE/%f %p'
recovery_target_time = '$TARGET_TIME'
recovery_target_action = 'promote'
EOF
touch "$PGDATA_RESTORE/recovery.signal"

pg_ctl -D "$PGDATA_RESTORE" -w start >/tmp/restore.log 2>&1 || {
  cat /tmp/restore.log >&2
  exit 1
}

BEFORE=$(psql -h "$SOCKET" -p "$PORT" -d postgres -t -A -c "SELECT count(*) FROM t WHERE id = 1")
AFTER=$(psql -h "$SOCKET" -p "$PORT" -d postgres -t -A -c "SELECT count(*) FROM t WHERE id = 2")

pg_ctl -D "$PGDATA_RESTORE" -m fast -w stop >/dev/null

if [ "$BEFORE" != "1" ]; then
  echo "FAIL: row 1 (before target) missing after recovery"
  exit 1
fi
if [ "$AFTER" != "0" ]; then
  echo "FAIL: row 2 (after target) present after recovery"
  exit 1
fi

echo "PITR_OK before=$BEFORE after=$AFTER"
