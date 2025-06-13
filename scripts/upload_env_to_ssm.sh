#!/bin/bash

# === Configuration ===
ENV_FILE=".env.deploy"
PREFIX="/weather-api"
REGION="us-east-1"
PROFILE="default"

# === Check if .env file exists ===
if [ ! -f "$ENV_FILE" ]; then
  echo "❌ File $ENV_FILE not found"
  exit 1
fi

echo "📤 Uploading parameters from $ENV_FILE to AWS SSM (region: $REGION, profile: $PROFILE)"
echo

# === Read and upload each variable ===
while IFS='=' read -r key value || [[ -n "$key" ]]; do
  # Skip comments and empty lines
  [[ "$key" =~ ^#.*$ || -z "$key" ]] && continue

  # Trim whitespace
  key=$(echo "$key" | xargs)
  value=$(echo "$value" | xargs)
  param_name="$PREFIX/$key"

  # Delete existing parameter (if any)
  echo "🗑️  Deleting $param_name (if exists)"
  aws ssm delete-parameter \
    --name "$param_name" \
    --region "$REGION" \
    --profile "$PROFILE" 2>/dev/null

  # Upload as fresh version 1
  echo "➕ Creating $param_name as version 1"
  aws ssm put-parameter \
    --name "$param_name" \
    --value "$value" \
    --type "SecureString" \
    --region "$REGION" \
    --profile "$PROFILE"
done < "$ENV_FILE"

echo
echo "✅ All parameters uploaded fresh as version 1."