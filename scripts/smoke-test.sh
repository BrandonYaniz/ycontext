#!/bin/sh
set -eu

root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
tmp=$(mktemp -d /tmp/ycontext-smoke.XXXXXX)
socket="$tmp/ycontextd.sock"
config="$tmp/ycontext.yaml"
daemon_log="$tmp/ycontextd.log"
daemon_pid=

cleanup() {
	if [ -n "$daemon_pid" ] && kill -0 "$daemon_pid" 2>/dev/null; then
		kill "$daemon_pid"
		wait "$daemon_pid" || true
	fi
	rm -rf "$tmp"
}
trap cleanup EXIT INT TERM

cat >"$config" <<EOF
server:
  socket_path: "$socket"
store:
  database_path: "$tmp/data/ycontext.db"
  document_path: "$tmp/documents"
llm:
  provider: "yllmd"
  socket_path: "$tmp/yllmd.sock"
  model: "local-default"
jobs:
  workers: 1
  idle_validation: true
EOF

printf 'Call me Ishmael.\nSome years ago never mind how long precisely.\n' >"$tmp/source.txt"

go build -o "$tmp/ycontext" "$root/cmd/ycontext"
go build -o "$tmp/ycontextd" "$root/cmd/ycontextd"

start_daemon() {
	"$tmp/ycontextd" -config "$config" >"$daemon_log" 2>&1 &
	daemon_pid=$!

	i=0
	while [ ! -S "$socket" ]; do
		if ! kill -0 "$daemon_pid" 2>/dev/null; then
			cat "$daemon_log"
			return 1
		fi
		i=$((i + 1))
		if [ "$i" -ge 100 ]; then
			cat "$daemon_log"
			return 1
		fi
		sleep 0.05
	done
}

stop_daemon() {
	kill "$daemon_pid"
	wait "$daemon_pid"
	daemon_pid=
	if [ -e "$socket" ]; then
		echo "socket was not removed after daemon shutdown" >&2
		return 1
	fi
}

start_daemon

status=$("$tmp/ycontext" -config "$config" status)
printf '%s\n' "$status" | grep -q '^ready: true$'

workspace=$("$tmp/ycontext" -config "$config" workspace create default | awk '/^workspace_id:/ {print $2}')
test -n "$workspace"

corpus=$("$tmp/ycontext" -config "$config" corpus create "$workspace" smoke | awk '/^corpus_id:/ {print $2}')
test -n "$corpus"

source=$("$tmp/ycontext" -config "$config" source add-text "$corpus" source.txt "$tmp/source.txt" | awk '/^source_id:/ {print $2}')
test -n "$source"

ingest=$("$tmp/ycontext" -config "$config" ingest start "$source" 4)
printf '%s\n' "$ingest" | grep -q '^chunks: 3$'

nodes=$("$tmp/ycontext" -config "$config" node list "$source")
test "$(printf '%s\n' "$nodes" | wc -l | tr -d ' ')" -eq 3
printf '%s\n' "$nodes" | grep -q 'rough_chunk'

stop_daemon
start_daemon

sources=$("$tmp/ycontext" -config "$config" source list "$corpus")
printf '%s\n' "$sources" | grep -q "$source"

nodes=$("$tmp/ycontext" -config "$config" node list "$source")
test "$(printf '%s\n' "$nodes" | wc -l | tr -d ' ')" -eq 3

stop_daemon
echo "smoke test passed"
