# Format all Go files with goimports and gofumpt
fmt:
	goimports -w .
	gofumpt -w .

# Run golangci-lint static code analysis
lint:
	golangci-lint run ./...

# Run all tests with verbose output
test:
	go test -v ./...

# Run tests quietly (remove Gin/debug noise)
test-quiet:
	@echo "==> Running tests quietly..."
	@go test ./... -v 2>&1 | grep -v -e '^\[GIN\]' -e 'record not found' -e '^=== RUN'

# Run all checks: formatting, linting and tests
check: fmt lint test-quiet

# Run locally with Docker Compose
run:
	docker-compose up --build

# ECS / CDK Deployment Settings
.PHONY: ecs cdk deploy fmt lint test test-quiet check run

ECR_URI=273354659544.dkr.ecr.us-east-1.amazonaws.com/weather-api
IMAGE_NAME=weather-api
PLATFORM=linux/amd64
CDK_DIR=cdk

# Build + push Docker image and redeploy to ECS
ecs:
	@echo "🐳 Building Docker image for $(PLATFORM)..."
	docker buildx build --platform=$(PLATFORM) --output=type=docker -t $(IMAGE_NAME) .

	@echo "🔑 Logging in to Amazon ECR..."
	aws ecr get-login-password --region $(REGION) | docker login --username AWS --password-stdin $(ECR_URI)

	@echo "🔁 Tagging image..."
	docker tag $(IMAGE_NAME):latest $(ECR_URI):latest

	@echo "🚀 Pushing image to ECR..."
	docker push $(ECR_URI):latest

	@echo "📦 Redeploying ECS service..."
	@./scripts/redeploy_ecs.sh

	@echo "✅ ECS redeployment complete."

# Deploy CDK stack
cdk:
	@echo "🚀 Deploying CDK stack from $(CDK_DIR)/ ..."
	cd $(CDK_DIR) && \
		if [ -f .venv/bin/activate ]; then \
			source .venv/bin/activate && \
			echo '🟢 Activated virtualenv' && \
			cdk deploy --require-approval never; \
		else \
			echo '❌ No virtualenv found in $(CDK_DIR). Please run: python3 -m venv .venv && source .venv/bin/activate && pip install -r requirements.txt'; \
			exit 1; \
		fi

# Full deploy: ECS image + CDK stack
deploy: ecs cdk
	@echo "✅ Full deployment complete."