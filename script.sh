# Inside rename.sh:
f="$1"; [ -e "$f" ] && mv -n -- "$f" "$(dirname -- "$f")/$(basename -- "$f" | tr -d '[:space:]')"