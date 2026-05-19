# swim-signup Makefile

.PHONY: build zip deploy dev-backend dev-frontend

# ── Backend ───────────────────────────────────────────────────────────────────

# Build the Go Lambda binary (ARM64 for Graviton, change to amd64 if preferred)
build:
	cd backend && \
	  GOOS=linux GOARCH=arm64 CGO_ENABLED=0 \
	  go build -tags lambda.norpc -o bootstrap ./cmd/lambda
	@echo "✓ Built backend/bootstrap"

# Zip for Lambda deployment
zip: build
	cd backend && zip lambda.zip bootstrap
	@echo "✓ Created backend/lambda.zip"

# Run the backend locally (requires AWS creds + env vars)
dev-backend:
	cd backend && \
	  PORT=8080 \
	  GOOGLE_CALENDAR_ID="$${GOOGLE_CALENDAR_ID}" \
	  GOOGLE_CREDENTIALS_JSON="$${GOOGLE_CREDENTIALS_JSON}" \
	  go run ./cmd/lambda

# ── Frontend ──────────────────────────────────────────────────────────────────

# Install frontend deps
frontend-install:
	cd frontend && npm install

# Run the frontend dev server (proxies API to localhost:8080)
dev-frontend:
	cd frontend && npm run dev

# Build the frontend for production
frontend-build:
	cd frontend && VITE_API_URL=$${API_URL} npm run build

# ── Infrastructure ────────────────────────────────────────────────────────────

# First-time Terraform init
infra-init:
	cd infra && terraform init

# Plan infra changes
infra-plan:
	cd infra && terraform plan \
	  -var="google_calendar_id=$${GOOGLE_CALENDAR_ID}"

# Deploy everything: build zip, then apply infra
deploy: zip
	cd infra && terraform apply \
	  -var="google_calendar_id=$${GOOGLE_CALENDAR_ID}" \
	  -var="google_credentials_json=$${GOOGLE_CREDENTIALS_JSON}" \
	  -auto-approve
	@echo ""
	@echo "✓ Deployed. API URL:"
	@cd infra && terraform output -raw api_url

# Update just the Lambda function code (faster than full apply)
update-lambda: zip
	aws lambda update-function-code \
	  --function-name swim-signup-api \
	  --zip-file fileb://backend/lambda.zip \
	  --architectures arm64