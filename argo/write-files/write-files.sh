#!/bin/sh

set -e

BASE_URL="$1"
RECORD_ID="$2"
FILES_DIR="/files"

if [ -z "$BASE_URL" ] || [ -z "$RECORD_ID" ]; then
  echo "Usage: $0 <base_url> <record_id>"
  exit 1
fi

if [ ! -d "$FILES_DIR" ]; then
  echo "Directory $FILES_DIR does not exist."
  exit 1
fi

for FILE_PATH in "$FILES_DIR"/*; do
  FILE_NAME=$(basename "$FILE_PATH")
  echo "Uploading file: $FILE_NAME"

  echo "Registering metadata"
  curl -sSf -X POST "${BASE_URL}/${RECORD_ID}/draft/files" \
    -H "Content-Type: application/json" \
    -d "{\"key\": \"${FILE_NAME}\"}" || { echo "Failed to register $FILE_NAME"; exit 1; }

  echo "Uploading content"
  curl -sSf -X POST "${BASE_URL}/${RECORD_ID}/draft/files/${FILE_NAME}/content" \
    --data-binary "@${FILE_PATH}" || { echo "Failed to upload content for $FILE_NAME"; exit 1; }

  echo "Committing file"
  curl -sSf -X POST "${BASE_URL}/${RECORD_ID}/draft/files/${FILE_NAME}/commit" \
    -H "Content-Length: 0" || { echo "Failed to commit $FILE_NAME"; exit 1; }

  echo "$FILE_NAME uploaded and committed successfully"
done

echo "All files uploaded successfully."
exit 0
