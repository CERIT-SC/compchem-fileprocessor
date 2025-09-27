#!/bin/sh

set -e

BASE_URL="$1"
WORKFLOW_NAME="$2"
SECRET_KEY="$3"

if [ -z "$BASE_URL" ] || [ -z "$WORKFLOW_NAME" ] || [ -z "$SECRET_KEY" ]; then
  echo "Usage: $0 <base_url> <workflow_name> <secret_key>"
  exit 1
fi

echo "Deleting context for workflow: $WORKFLOW_NAME"

curl -f -k -H "Host: localhost" -X DELETE "${BASE_URL}/workflows/${WORKFLOW_NAME}/context?secret_key=${SECRET_KEY}" || {
  echo "Failed to delete context for workflow $WORKFLOW_NAME"
  exit 1
}

echo "Context deleted successfully for workflow: $WORKFLOW_NAME"
exit 0