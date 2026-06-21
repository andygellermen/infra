#!/usr/bin/env bash

ghost_registry_manifest_accept_header='application/vnd.oci.image.index.v1+json, application/vnd.docker.distribution.manifest.list.v2+json, application/vnd.docker.distribution.manifest.v2+json'

ghost_registry_token() {
  curl -fsSL 'https://auth.docker.io/token?service=registry.docker.io&scope=repository:library/ghost:pull' \
    | sed -n 's/.*"token":"\([^"]*\)".*/\1/p'
}

ghost_image_tag_status() {
  local tag="${1:?missing ghost tag}"
  local token

  token="$(ghost_registry_token)" || return 2
  [[ -n "$token" ]] || return 2

  curl -sSIL -o /dev/null -w '%{http_code}' \
    -H "Authorization: Bearer ${token}" \
    -H "Accept: ${ghost_registry_manifest_accept_header}" \
    "https://registry-1.docker.io/v2/library/ghost/manifests/${tag}" \
    || true
}

validate_ghost_image_tag_or_die() {
  local tag="${1:?missing ghost tag}"
  local context="${2:-Ghost-Version}"
  local status_code

  if [[ "${INFRA_GHOST_TAG_PREVALIDATED:-}" == "$tag" ]]; then
    return 0
  fi

  status_code="$(ghost_image_tag_status "$tag")"
  case "$status_code" in
    200)
      export INFRA_GHOST_TAG_PREVALIDATED="$tag"
      return 0
      ;;
    404)
      echo "❌ Fehler: ${context} '${tag}' ist aktuell nicht als offizielles Docker-Image ghost:${tag} verfuegbar." >&2
      echo "   Hinweis: Ein Ghost-Release kann bereits existieren, waehrend das Docker-Tag noch nicht publiziert ist." >&2
      return 1
      ;;
    *)
      echo "❌ Fehler: Das Docker-Image ghost:${tag} konnte nicht gegen Docker Hub verifiziert werden (HTTP ${status_code:-unbekannt})." >&2
      echo "   Bitte Netzwerkzugang zu Docker Hub pruefen oder den Vorgang spaeter erneut versuchen." >&2
      return 1
      ;;
  esac
}
