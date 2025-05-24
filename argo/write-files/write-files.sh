#!/bin/sh

set -e

BASE_URL="$1"
RECORD_ID="$2"
WORKFLOW_NAME="$3"
TASK_DISCRIMINATOR="$4"
FILES_DIR="/input"

if [ -z "$BASE_URL" ] || [ -z "$RECORD_ID" ]; then
  echo "Usage: $0 <base_url> <record_id>"
  exit 1
fi

if [ ! -d "$FILES_DIR" ]; then
  echo "Directory $FILES_DIR is not present."
  exit 1
fi

for FILE_PATH in "$FILES_DIR"/*; do
  FILE_NAME=$(basename "$FILE_PATH")-$JOB_DISCRIMINATOR-$WORKFLOW_NAME
  echo "Uploading file: $FILE_NAME"

  # TODO: temporary curl headers, -k -H
  # TODO: what is the purpose here? do we delete the old file? or keep creating new ones?
  echo "Registering metadata"
  curl -f -k -H "Host: localhost" -X POST "${BASE_URL}/${RECORD_ID}/draft/files" \
    -H "Content-Type: application/json" \
    -d "[{\"key\": \"${FILE_NAME}\"}]" || { echo "Failed to register $FILE_NAME"; exit 1; }

  echo "\n"

  echo "Uploading content"
    curl -f -k -H "Host: localhost" -H "Content-Type: application/octet-stream" -X PUT "${BASE_URL}/${RECORD_ID}/draft/files/${FILE_NAME}/content" \
    --data-binary "@${FILE_PATH}" || { echo "Failed to upload content for $FILE_NAME"; exit 1; }

  echo "\n"

  echo "Committing file"
  curl -f -k -H "Host: localhost" -X POST "${BASE_URL}/${RECORD_ID}/draft/files/${FILE_NAME}/commit" || { echo "Failed to commit $FILE_NAME"; exit 1; }

  echo "$FILE_NAME uploaded and committed successfully"
done

echo "All files uploaded successfully."
exit 0
