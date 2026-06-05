# WildLens Query API

Go API for querying WildLens media metadata.

## Local Memory Mode

```bash
cd backend/query-api
APP_REPOSITORY=memory go run ./cmd/server
```

The local memory mode does not require AWS credentials.

## Local Tests

```bash
cd backend/query-api
go fmt ./...
go mod tidy
go test -v ./...
```

## Cloud Run Deployment

Run from the repository root:

```bash
chmod +x backend/query-api/scripts/deploy-cloud-run.sh
./backend/query-api/scripts/deploy-cloud-run.sh
```

The deploy script silently prompts for temporary AWS Academy credentials and writes them to Google Cloud Secret Manager. It does not print the credential values.

Before running the script, make sure Google Cloud CLI is installed, you are logged in, and a project is selected:

```bash
gcloud auth login
gcloud config set project <PROJECT_ID>
```

## Refresh AWS Academy Credentials

AWS Academy Lab credentials change after the lab restarts. Refresh the Cloud Run secrets without changing code:

```bash
chmod +x backend/query-api/scripts/refresh-aws-secrets.sh
./backend/query-api/scripts/refresh-aws-secrets.sh
```

The refresh script adds new Secret Manager versions and updates Cloud Run to use the latest versions.

## Security

- Do not commit real AWS credentials.
- Do not write credentials into `.env`.
- Do not paste credentials into chat, group chat, or Codex.
- After AWS Academy credentials expire, run the refresh script.
