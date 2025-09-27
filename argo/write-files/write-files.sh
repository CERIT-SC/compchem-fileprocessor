#!/bin/sh

set -e

BASE_URL="$1"
RECORD_ID="$2"
WORKFLOW_NAME="$3"
TASK_DISCRIMINATOR="$4"
SECRET_KEY="$5"
FILES_DIR="/input"

if [ -z "$BASE_URL" ] || [ -z "$RECORD_ID" ]; then
  echo "Usage: $0 <base_url> <record_id> <workflow_name> <task_discriminator> <secret_key>"
  exit 1
fi

if [ ! -d "$FILES_DIR" ]; then
  echo "Directory $FILES_DIR is not present."
  exit 1
fi

for FILE_PATH in "$FILES_DIR"/*; do
  FILE_NAME=$WORKFLOW_NAME-$TASK_DISCRIMINATOR-$(basename "$FILE_PATH")
  echo "Uploading file: $FILE_NAME"

  echo "Uploading content"
    curl -f -k -H "Host: localhost" -H "Content-Type: application/octet-stream" -X PUT "${BASE_URL}/${RECORD_ID}/draft/files/${FILE_NAME}/workflow-commit?secret_key=${SECRET_KEY}" \
    --data-binary "@${FILE_PATH}" || { echo "Failed to upload content for $FILE_NAME"; exit 1; }

done

echo "All files uploaded successfully."
exit 0
