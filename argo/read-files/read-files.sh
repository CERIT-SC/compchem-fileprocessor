#!/bin/bash

if [ "$#" -lt 4 ]; then
  echo "Usage: $0 <url> <record_id> <secret_key> <file_id1> [file_id2] [...]"
  exit 1
fi

URL="$1" # https://localhost:5000/api/experiments
RECORD_ID="$2" # ew6jd-p8175
SECRET_KEY="$3" # workflow secret key
shift 3

DOWNLOAD_DIR="/output"
mkdir -p $DOWNLOAD_DIR

# Create JSON array of file keys
FILE_KEYS_JSON=""
for FILE_ID in "$@"; do
  if [ -z "$FILE_KEYS_JSON" ]; then
    FILE_KEYS_JSON="\"$FILE_ID\""
  else
    FILE_KEYS_JSON="$FILE_KEYS_JSON, \"$FILE_ID\""
  fi
done

# Create the JSON payload
JSON_PAYLOAD="{
  \"record_id\": \"$RECORD_ID\",
  \"secret_key\": \"$SECRET_KEY\",
  \"file_keys\": [$FILE_KEYS_JSON]
}"

echo "Requesting signed URLs for files: $@"
echo "JSON payload: $JSON_PAYLOAD"

# Request signed URLs from the new workflow files endpoint
READ_URL="${URL}/workflow_files/read"
RESPONSE=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Host: localhost:5000" \
  --insecure \
  -d "$JSON_PAYLOAD" \
  "$READ_URL")

echo "Response: $RESPONSE"

# Check if request was successful
if echo "$RESPONSE" | grep -q '"error"'; then
  echo "Error getting signed URLs: $RESPONSE"
  exit 1
fi

# Download each file using the signed URLs
for FILE_ID in "$@"; do
  OUTPUT_FILE="${DOWNLOAD_DIR}/${FILE_ID}"

  # Extract signed URL for this file from JSON response
  SIGNED_URL=$(echo "$RESPONSE" | jq -r ".files.\"$FILE_ID\".signed_url // empty")

  if [ -z "$SIGNED_URL" ] || [ "$SIGNED_URL" = "null" ]; then
    echo "No signed URL found for file: $FILE_ID"
    # Check for error in response
    ERROR_MSG=$(echo "$RESPONSE" | jq -r ".errors.\"$FILE_ID\" // \"Unknown error\"")
    echo "Error: $ERROR_MSG"
    continue
  fi

  echo "Downloading $FILE_ID from: $SIGNED_URL"
  echo "Saving to: $OUTPUT_FILE"

  # Download directly from signed URL
  wget --no-check-certificate -O "$OUTPUT_FILE" "$SIGNED_URL" || {
    echo "Failed to download $FILE_ID from signed URL";
    continue;
  }

  echo "Successfully downloaded: $FILE_ID"
done

echo "Download complete"
