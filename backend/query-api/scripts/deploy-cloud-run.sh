#!/usr/bin/env bash
set -euo pipefail

SERVICE_NAME="${SERVICE_NAME:-wildlens-query-api}"
REGION="${REGION:-australia-southeast1}"
SOURCE_DIR="${SOURCE_DIR:-backend/query-api}"
AWS_REGION="${AWS_REGION:-us-east-1}"
DYNAMODB_TABLE_NAME="${DYNAMODB_TABLE_NAME:-fit5225-wildlife-media-metadata}"
RUNNER_SERVICE_ACCOUNT="${RUNNER_SERVICE_ACCOUNT:-wildlens-query-api-runner}"

AWS_ACCESS_KEY_ID_SECRET="wildlens-aws-access-key-id"
AWS_SECRET_ACCESS_KEY_SECRET="wildlens-aws-secret-access-key"
AWS_SESSION_TOKEN_SECRET="wildlens-aws-session-token"

cleanup_credentials() {
  unset AWS_ACCESS_KEY_ID || true
  unset AWS_SECRET_ACCESS_KEY || true
  unset AWS_SESSION_TOKEN || true
}

trap cleanup_credentials EXIT

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
QUERY_API_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${QUERY_API_DIR}/../.." && pwd)"
cd "${REPO_ROOT}"

if ! command -v gcloud >/dev/null 2>&1; then
  echo "Google Cloud CLI is required. Install it first: https://cloud.google.com/sdk/docs/install" >&2
  exit 1
fi

ACTIVE_ACCOUNT="$(
  gcloud auth list --filter=status:ACTIVE --format="value(account)" 2>/dev/null || true
)"

if [[ -z "${ACTIVE_ACCOUNT}" ]]; then
  echo "No active gcloud account. Run: gcloud auth login" >&2
  exit 1
fi

PROJECT_ID="$(
  gcloud config get-value project 2>/dev/null || true
)"

if [[ -z "${PROJECT_ID}" || "${PROJECT_ID}" == "(unset)" ]]; then
  echo "No active GCP project. Run: gcloud config set project <PROJECT_ID>" >&2
  exit 1
fi

RUNNER_SERVICE_ACCOUNT_EMAIL="${RUNNER_SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com"

gcloud services enable \
  run.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  secretmanager.googleapis.com

if ! gcloud iam service-accounts describe "${RUNNER_SERVICE_ACCOUNT_EMAIL}" >/dev/null 2>&1; then
  gcloud iam service-accounts create "${RUNNER_SERVICE_ACCOUNT}" \
    --display-name="WildLens Query API Runner"
fi

read -r -s -p "AWS_ACCESS_KEY_ID: " AWS_ACCESS_KEY_ID
echo
read -r -s -p "AWS_SECRET_ACCESS_KEY: " AWS_SECRET_ACCESS_KEY
echo
read -r -s -p "AWS_SESSION_TOKEN: " AWS_SESSION_TOKEN
echo

if [[ -z "${AWS_ACCESS_KEY_ID}" || -z "${AWS_SECRET_ACCESS_KEY}" || -z "${AWS_SESSION_TOKEN}" ]]; then
  echo "All three AWS credential values are required." >&2
  exit 1
fi

ensure_secret_exists() {
  local secret_name="$1"

  if ! gcloud secrets describe "${secret_name}" >/dev/null 2>&1; then
    gcloud secrets create "${secret_name}" --replication-policy=automatic
  fi
}

ensure_secret_exists "${AWS_ACCESS_KEY_ID_SECRET}"
ensure_secret_exists "${AWS_SECRET_ACCESS_KEY_SECRET}"
ensure_secret_exists "${AWS_SESSION_TOKEN_SECRET}"

printf '%s' "${AWS_ACCESS_KEY_ID}" |
  gcloud secrets versions add "${AWS_ACCESS_KEY_ID_SECRET}" --data-file=-

printf '%s' "${AWS_SECRET_ACCESS_KEY}" |
  gcloud secrets versions add "${AWS_SECRET_ACCESS_KEY_SECRET}" --data-file=-

printf '%s' "${AWS_SESSION_TOKEN}" |
  gcloud secrets versions add "${AWS_SESSION_TOKEN_SECRET}" --data-file=-

cleanup_credentials

grant_secret_access() {
  local secret_name="$1"

  gcloud secrets add-iam-policy-binding "${secret_name}" \
    --member="serviceAccount:${RUNNER_SERVICE_ACCOUNT_EMAIL}" \
    --role="roles/secretmanager.secretAccessor"
}

grant_secret_access "${AWS_ACCESS_KEY_ID_SECRET}"
grant_secret_access "${AWS_SECRET_ACCESS_KEY_SECRET}"
grant_secret_access "${AWS_SESSION_TOKEN_SECRET}"

gcloud run deploy "${SERVICE_NAME}" \
  --source "${SOURCE_DIR}" \
  --region "${REGION}" \
  --platform managed \
  --allow-unauthenticated \
  --service-account "${RUNNER_SERVICE_ACCOUNT_EMAIL}" \
  --set-env-vars "APP_REPOSITORY=dynamodb,AWS_REGION=${AWS_REGION},DYNAMODB_TABLE_NAME=${DYNAMODB_TABLE_NAME}" \
  --update-secrets "AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID_SECRET}:latest,AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY_SECRET}:latest,AWS_SESSION_TOKEN=${AWS_SESSION_TOKEN_SECRET}:latest"

SERVICE_URL="$(
  gcloud run services describe "${SERVICE_NAME}" \
    --region "${REGION}" \
    --format='value(status.url)'
)"

echo "Cloud Run URL: ${SERVICE_URL}"

curl --fail --show-error "${SERVICE_URL}/health"

if ! curl --fail --show-error "${SERVICE_URL}/api/v1/observations?limit=5"; then
  echo "Observation smoke test failed. Check:" >&2
  echo "- AWS Academy Credentials 是否过期" >&2
  echo "- APP_REPOSITORY 是否为 dynamodb" >&2
  echo "- AWS_REGION 是否为 us-east-1" >&2
  echo "- DynamoDB Table 名是否正确" >&2
  echo "- Cloud Run logs" >&2
  exit 1
fi
