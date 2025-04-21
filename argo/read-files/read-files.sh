#!/bin/bash

if [ "$#" -lt 3 ]; then
  echo "Usage: $0 <url> <file_id1> [file_id2] [...]"
  exit 1
fi

URL="$1" # https://localhost:5000/api/experiments
shift
RECORD_ID="$2" # ew6jd-p8175
shift 

DOWNLOAD_DIR="/files"
mkdir -p $DOWNLOAD_DIR

for FILE_ID in "$@"; do
  DOWNLOAD_URL="${URL}/${RECORD_ID$}/draft/files/${FILE_ID}/content"
  OUTPUT_FILE="${DOWNLOAD_DIR}/${FILE_ID}"

  echo "Downloading from: $DOWNLOAD_URL"
  echo "Saving to: $OUTPUT_FILE"

  wget --no-check-certificate -H "Host: localhost" "$DOWNLOAD_URL" -O "$OUTPUT_FILE"

  if [ $? -eq 0 ]; then
    echo "Downloaded $FILE_ID successfully."
  else
    echo "Failed to download $FILE_ID"
    exit 1
  fi

  echo
done

exit 0
