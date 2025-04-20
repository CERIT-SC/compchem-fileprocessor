#!/bin/bash

if [ "$#" -lt 2 ]; then
  echo "Usage: $0 <url> <file_id1> [file_id2] [...]"
  exit 1
fi

URL="$1"
shift 

DOWNLOAD_DIR="/files"
mkdir -p $DOWNLOAD_DIR

for FILE_ID in "$@"; do
  DOWNLOAD_URL="${BASE_URL}/${FILE_ID}/download"
  OUTPUT_FILE="${DOWNLOAD_DIR}/${FILE_ID}"

  echo "Downloading from: $DOWNLOAD_URL"
  echo "Saving to: $OUTPUT_FILE"

  curl -sSL "$DOWNLOAD_URL" -o "$OUTPUT_FILE"

  if [ $? -eq 0 ]; then
    echo "Downloaded $FILE_ID successfully."
  else
    echo "Failed to download $FILE_ID"
    exit 1
  fi

  echo
done

exit 0
