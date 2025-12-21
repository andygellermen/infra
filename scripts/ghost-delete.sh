DOMAIN="$1"
CONTAINER_NAME="ghost-${DOMAIN//./-}"

echo "🔎 Prüfe Docker-Container: $CONTAINER_NAME"

if docker ps -a --format '{{.Names}}' | grep -qx "$CONTAINER_NAME"; then
  echo "🛑 Stoppe Ghost-Container (falls laufend)..."
  docker stop "$CONTAINER_NAME" >/dev/null 2>&1 || true

  echo "🗑️  Entferne Ghost-Container..."
  docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true

  echo "✅ Container vollständig entfernt"
else
  echo "ℹ️  Kein Ghost-Container gefunden"
fi

HOSTVARS_FILE="./ansible/hostvars/${DOMAIN}.yml"

if [[ -f "$HOSTVARS_FILE" ]]; then
  rm -f "$HOSTVARS_FILE"
  echo "🗑️  Hostvars gelöscht: $HOSTVARS_FILE"
else
  echo "ℹ️  Keine Hostvars-Datei gefunden für $DOMAIN"
fi


echo "⏳ Warte kurz, damit Docker Ressourcen freigibt..."
sleep 2
