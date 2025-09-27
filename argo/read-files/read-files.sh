#!/bin/bash

if [ "$#" -lt 3 ]; then
  echo "Usage: $0 <url> <secret-key> <file_id1> [file_id2] [...]"
  exit 1
fi

URL="$1" # https://localhost:5000/api/experiments
shift
SECRET="$1"
shift
RECORD_ID="$1" # ew6jd-p8175
shift

DOWNLOAD_DIR="/output"
mkdir -p $DOWNLOAD_DIR

for FILE_ID in "$@"; do
  DOWNLOAD_URL="${URL}/${RECORD_ID}/draft/files/${FILE_ID}/workflow-content?secret_key=${SECRET}"
  OUTPUT_FILE="${DOWNLOAD_DIR}/${FILE_ID}"
  echo "Downloading from: $DOWNLOAD_URL"
  echo "Saving to: $OUTPUT_FILE"

  REDIRECT_OUTPUT=$(wget --no-check-certificate --header="Host: localhost:5000" --max-redirect=0 "$DOWNLOAD_URL" -O "$OUTPUT_FILE" 2>&1)
  REDIRECT_URL=$(echo "$REDIRECT_OUTPUT" | grep -o "Location:.*" | cut -d' ' -f2- | sed 's/ \[following\]$//' | sed 's/127.0.0.1/172.22.0.3/')

  if [ -z "$REDIRECT_URL" ]; then
    echo "Failed to get redirect URL from ${DOWNLOAD_URL}"
    echo "$REDIRECT_OUTPUT"
    exit 1
  fi

  echo "Redirected to: $REDIRECT_URL"

  wget --header="Host: 127.0.0.1:9000" -O "$OUTPUT_FILE" "$REDIRECT_URL" || {
    echo "Failed to download from redirect URL";
    exit 1;
  }
done
