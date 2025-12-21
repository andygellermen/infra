#!/bin/bash

# delete-hostvars.sh
# Löscht Hostvars-Datei einer Domain (inkl. IDN)

set -e

if [[ -z "$1" ]]; then
    echo "Usage: $0 <domain>"
    exit 1
fi

if ! command -v idn >/dev/null 2>&1; then
    echo "Fehler: Das 'idn'-Tool ist nicht installiert. Bitte mit 'sudo apt install idn' nachholen."
    exit 1
fi

domain=$(idn --quiet "$1")
file="./ansible/hostvars/${domain}.yml"

if [[ -f "$file" ]]; then
    echo "🗑️  Lösche Datei: $file"
    rm "$file"
else
    echo "⚠️  Datei nicht gefunden: $file"
fi
