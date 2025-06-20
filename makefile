# Format all Go files
fmt:
	goimports -w .
	gofumpt -w .

# Run static code analysis
lint:
	golangci-lint run --fix ./...

# Run all tests (unit + integration)
test:
	gotestsum -- -count=1 -tags=integration ./...

# Run only unit tests (all tests that do NOT have the 'integration' build tag)
test-unit:
	gotestsum -- -count=1 ./...

# Run only integration tests (tests with //go:build integration)
test-integration:
	gotestsum -- -count=1 -tags=integration ./internal/integration/...


e2e-up:
	docker compose up -d --quiet-pull

e2e-test:
	go test -tags=e2e ./test/e2e -v

e2e-down:
	docker compose down -v

e2e: e2e-up
	sleep 5
	make e2e-test
	make e2e-down


# Run formatting, linting and tests
check: fmt lint test

# Run locally using Docker Compose
run:
	docker-compose up --build

# Build and deploy Docker image to Amazon ECS
ecs:
	@echo "Building Docker image for $(PLATFORM)..."
	docker buildx build \
		--platform=$(PLATFORM) \
		--output=type=docker \
		-t $(IMAGE_NAME) .

	@echo "Logging in to Amazon ECR..."
	aws ecr get-login-password --region $(REGION) | \
	docker login --username AWS --password-stdin $(ECR_URI)

	@echo "Tagging image..."
	docker tag $(IMAGE_NAME):latest $(ECR_URI):latest

	@echo "Pushing image to ECR..."
	docker push $(ECR_URI):latest

	@echo "Redeploying ECS service..."
	./scripts/redeploy_ecs.sh

	@echo "ECS redeployment complete."

# Deploy AWS CDK stack
cdk:
	@echo "Deploying CDK stack from $(CDK_DIR)/ ..."
	cd $(CDK_DIR) && \
	if [ -f .venv/bin/activate ]; then \
		source .venv/bin/activate && \
		cdk deploy --require-approval never; \
	else \
		echo "No virtualenv found. Please run:"; \
		echo "python3 -m venv .venv && source .venv/bin/activate && pip install -r requirements.txt"; \
		exit 1; \
	fi

# Full deployment: Docker image + CDK stack
deploy: ecs cdk
	@echo "Full deployment complete."

.PHONY: fmt lint test test-quiet check run ecs cdk deploy