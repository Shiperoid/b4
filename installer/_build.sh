#!/bin/sh
# Builds install.sh from parts

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
OUTPUT="${SCRIPT_DIR}/../install.sh"

PARTS="
header.sh
config.sh
colors.sh
utils.sh
system.sh
download.sh
service.sh
geodat.sh
core.sh
remove.sh
update.sh
sysinfo.sh
main.sh
"

>"$OUTPUT"

for part in $PARTS; do
    file="${SCRIPT_DIR}/${part}"
    if [ ! -f "$file" ]; then
        echo "Missing: $part" >&2
        exit 1
    fi

    if [ "$part" = "header.sh" ]; then
        cat "$file" >>"$OUTPUT"
    else
        tail -n +2 "$file" | sed '/./,$!d' >>"$OUTPUT"
    fi

    echo "" >>"$OUTPUT"
    echo "# --- END $part ---" >>"$OUTPUT"
    echo "" >>"$OUTPUT"
done

chmod +x "$OUTPUT"
echo "Built: $OUTPUT ($(wc -l <"$OUTPUT") lines)"
