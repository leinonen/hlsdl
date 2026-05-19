#!/bin/bash

set -euo pipefail

SEGMENTS_DIR="${1:-segments}"
OUTPUT="${2:-output.mp4}"

if [[ ! -d "$SEGMENTS_DIR" ]]; then
    echo "error: directory '$SEGMENTS_DIR' not found" >&2
    exit 1
fi

LIST=$(mktemp /tmp/ffmpeg-concat-XXXXXX.txt)
trap "rm -f $LIST" EXIT

# sort numerically by segment number
for f in $(ls "$SEGMENTS_DIR"/*.ts | sort -t- -k2 -n); do
    echo "file '$(realpath "$f")'" >> "$LIST"
done

COUNT=$(wc -l < "$LIST")
echo "assembling $COUNT segments → $OUTPUT"

ffmpeg -y -f concat -safe 0 -i "$LIST" -c copy "$OUTPUT"
echo "done: $OUTPUT"
