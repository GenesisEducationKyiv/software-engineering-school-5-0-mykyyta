# Non-Functional Requirements — Weather API

## 1. Performance
- API should respond within 200ms for most requests
- Email delivery tasks should not block HTTP handlers

## 2. Reliability
- Email confirmations and deliveries must be reliable (retry logic planned)
- Scheduled tasks must run without data loss

## 3. Security
- Token links must be signed using JWT
- Sensitive fields (e.g., token) are excluded from API responses

## 4. Maintainability
- Environment variables are stored in `.env` and managed in AWS SSM
- Application is containerized for easy deployment

## 5. Scalability
- System can scale horizontally using ECS
- PostgreSQL can be migrated to managed service (e.g., RDS) later

## 6. Observability
- Logs for email failures, confirmation status
- Health check endpoint for readiness/liveness