#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOSTVARS_DIR="$ROOT_DIR/ansible/hostvars"
GHOST_REDEPLOY="$ROOT_DIR/scripts/ghost-redeploy.sh"
WP_REDEPLOY="$ROOT_DIR/scripts/wp-redeploy.sh"
STATIC_REDEPLOY="$ROOT_DIR/scripts/static-redeploy.sh"

usage() {
  cat <<'USAGE'
Usage:
  ./scripts/redeploy-all-web.sh [--check-only] [--only=all|ghost|wp|static] [--parallel=<n>] [--continue-on-error]

Description:
  Redeployt alle Web-Container (Ghost + WordPress + Static) basierend auf Hostvars.
USAGE
}

die(){ echo "❌ $*" >&2; exit 1; }
info(){ echo "ℹ️  $*"; }
ok(){ echo "✅ $*"; }
warn(){ echo "⚠️  $*"; }

CHECK_ONLY=0
ONLY="all"
CONTINUE_ON_ERROR=0
PARALLEL=1

while [[ $# -gt 0 ]]; do
  case "$1" in
    --check-only) CHECK_ONLY=1; shift ;;
    --only=*) ONLY="${1#*=}"; shift ;;
    --parallel=*) PARALLEL="${1#*=}"; shift ;;
    --continue-on-error) CONTINUE_ON_ERROR=1; shift ;;
    --help|-h) usage; exit 0 ;;
    *) die "Unbekannte Option: $1" ;;
  esac
done

[[ "$ONLY" =~ ^(all|ghost|wp|static)$ ]] || die "--only muss all|ghost|wp|static sein."
[[ "$PARALLEL" =~ ^[0-9]+$ ]] || die "--parallel muss eine Zahl >= 1 sein."
(( PARALLEL >= 1 )) || die "--parallel muss >= 1 sein."
[[ -d "$HOSTVARS_DIR" ]] || die "Hostvars-Verzeichnis fehlt: $HOSTVARS_DIR"
[[ -x "$GHOST_REDEPLOY" ]] || die "Script nicht ausführbar: $GHOST_REDEPLOY"
[[ -x "$WP_REDEPLOY" ]] || die "Script nicht ausführbar: $WP_REDEPLOY"
[[ -x "$STATIC_REDEPLOY" ]] || die "Script nicht ausführbar: $STATIC_REDEPLOY"

mapfile -t HOSTVAR_FILES < <(find "$HOSTVARS_DIR" -maxdepth 1 -type f -name '*.yml' | sort)
[[ ${#HOSTVAR_FILES[@]} -gt 0 ]] || die "Keine Hostvars-Dateien in $HOSTVARS_DIR gefunden."

ghost_domains=()
wp_domains=()
static_domains=()

for file in "${HOSTVAR_FILES[@]}"; do
  domain="$(basename "$file" .yml)"
  grep -q '^ghost_domain_db:' "$file" && ghost_domains+=("$domain")
  grep -q '^wp_domain_db:' "$file" && wp_domains+=("$domain")
  grep -q '^static_enabled:[[:space:]]*true' "$file" && static_domains+=("$domain")
done

run_redeploy() {
  local kind="$1" domain="$2" script="$3"
  local cmd=("$script" "$domain")
  [[ "$CHECK_ONLY" -eq 1 ]] && cmd+=("--check-only")

  info "Starte ${kind}-Redeploy: ${domain}"
  if "${cmd[@]}"; then
    ok "${kind} erfolgreich: ${domain}"
    return 0
  fi

  warn "${kind} fehlgeschlagen: ${domain}"
  return 1
}

failed=()
declare -A pid_to_entry=()
pids=()

wait_for_pid() {
  local pid="$1" rc entry
  entry="${pid_to_entry[$pid]}"
  set +e
  wait "$pid"
  rc=$?
  set -e
  unset 'pid_to_entry[$pid]'

  if [[ "$rc" -ne 0 ]]; then
    failed+=("$entry")
    return 1
  fi
  return 0
}

run_parallel_batch() {
  local kind="$1" script="$2"
  shift 2
  local domains=("$@")
  local abort=0

  for d in "${domains[@]}"; do
    [[ "$abort" -eq 1 ]] && break
    (
      run_redeploy "$kind" "$d" "$script"
    ) &
    pid=$!
    pids+=("$pid")
    pid_to_entry["$pid"]="${kind,,}:$d"

    while [[ "${#pids[@]}" -ge "$PARALLEL" ]]; do
      sleep 0.2
      new_pids=()
      for p in "${pids[@]}"; do
        if kill -0 "$p" 2>/dev/null; then
          new_pids+=("$p")
        else
          if ! wait_for_pid "$p"; then
            [[ "$CONTINUE_ON_ERROR" -eq 1 ]] || abort=1
          fi
        fi
      done
      pids=("${new_pids[@]}")
      [[ "$abort" -eq 1 ]] && break
    done
  done

  for p in "${pids[@]}"; do
    if ! wait_for_pid "$p"; then
      [[ "$CONTINUE_ON_ERROR" -eq 1 ]] || abort=1
    fi
  done
  pids=()

  return "$abort"
}

if [[ "$ONLY" == "all" || "$ONLY" == "ghost" ]]; then
  info "Ghost-Domains: ${#ghost_domains[@]}"
  if [[ "$PARALLEL" -gt 1 ]]; then
    run_parallel_batch "Ghost" "$GHOST_REDEPLOY" "${ghost_domains[@]}" || [[ "$CONTINUE_ON_ERROR" -eq 1 ]] || exit 1
  else
    for d in "${ghost_domains[@]}"; do
      if ! run_redeploy "Ghost" "$d" "$GHOST_REDEPLOY"; then
        failed+=("ghost:${d}")
        [[ "$CONTINUE_ON_ERROR" -eq 1 ]] || exit 1
      fi
    done
  fi
fi

if [[ "$ONLY" == "all" || "$ONLY" == "wp" ]]; then
  info "WordPress-Domains: ${#wp_domains[@]}"
  if [[ "$PARALLEL" -gt 1 ]]; then
    run_parallel_batch "WordPress" "$WP_REDEPLOY" "${wp_domains[@]}" || [[ "$CONTINUE_ON_ERROR" -eq 1 ]] || exit 1
  else
    for d in "${wp_domains[@]}"; do
      if ! run_redeploy "WordPress" "$d" "$WP_REDEPLOY"; then
        failed+=("wp:${d}")
        [[ "$CONTINUE_ON_ERROR" -eq 1 ]] || exit 1
      fi
    done
  fi
fi

if [[ "$ONLY" == "all" || "$ONLY" == "static" ]]; then
  info "Static-Domains: ${#static_domains[@]}"
  if [[ ${#static_domains[@]} -gt 0 ]]; then
    info "Starte Shared Static-Redeploy"
    if ! "$STATIC_REDEPLOY" --all $([[ "$CHECK_ONLY" -eq 1 ]] && printf '%s' '--check-only'); then
      failed+=("static:shared")
      [[ "$CONTINUE_ON_ERROR" -eq 1 ]] || exit 1
    else
      ok "Static erfolgreich: shared"
    fi
  fi
fi

if [[ ${#failed[@]} -gt 0 ]]; then
  warn "Redeploy mit Fehlern beendet:"
  for entry in "${failed[@]}"; do
    warn "  - $entry"
  done
  exit 1
fi

ok "Alle gewünschten Redeploys abgeschlossen."
