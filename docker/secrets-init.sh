#!/bin/sh
# Generates this install's JWT secrets on first run so `docker compose up`
# needs no setup. An existing file is never regenerated.
set -e
f=/run/mandarin-secrets/secrets.env
fail() { echo "secrets-init: $*" >&2; exit 1; }
valid() { case "$1" in ''|*[!0-9a-f]*) return 1 ;; esac; [ "${#1}" -eq 64 ]; }
gen() { head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n'; }
if [ ! -s "$f" ]; then
  umask 077
  a=$(gen); b=$(gen)
  for v in "$a" "$b"; do
    valid "$v" || fail "could not generate a secret"
  done
  printf 'JWT_SECRET=%s\nJWT_REFRESH_SECRET=%s\n' "$a" "$b" > "$f.tmp"
  chown 65532:0 "$f.tmp" || fail "could not set the owner of $f.tmp"
  chmod 400 "$f.tmp" || fail "could not set the mode of $f.tmp"
  mv "$f.tmp" "$f"
fi
n=0
while IFS='=' read -r k v; do
  case "$k" in
    JWT_SECRET|JWT_REFRESH_SECRET) ;;
    *) fail "$f has an unexpected line ($k)" ;;
  esac
  valid "$v" || fail "$f has an invalid $k. Remove this project's secrets volume to generate new ones."
  n=$((n + 1))
done < "$f"
[ "$n" -eq 2 ] || fail "$f should hold 2 secrets but holds $n"
chown 65532:0 "$f" || fail "could not set the owner of $f"
chmod 400 "$f" || fail "could not set the mode of $f"
