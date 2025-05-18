# weatherApi

A lightweight REST API built with Go and Gin for retrieving weather data, managing city subscriptions, and sending email notifications. Includes JWT-based authentication, email confirmation, and IaC setup via AWS CDK.

## Tech Stack

* **Language:** Go
* **Framework:** Gin
* **Database:** PostgreSQL
* **ORM:** GORM
* **Authentication:** JWT
* **Email Service:** SendGrid
* **Infrastructure as Code:** AWS CDK (Python)
* **Containerization:** Docker
* **Testing:** `testing`, `httptest`, `stretchr/testify`
* **API Docs:** Swagger (OpenAPI)

## Project Structure

```
.
├── cmd/server/main.go           # Entry point
├── config/config.go             # App configuration
├── internal/                    # Core logic
│   ├── api/                     # Handlers & tests
│   ├── db/                      # DB connection
│   └── model/                   # Data models
├── pkg/                         # Shared utilities
│   ├── email/                   # SendGrid integration
│   ├── jwtutil/                 # JWT utilities
│   ├── scheduler/               # Periodic tasks
│   └── weatherapi/              # Weather API client
├── templates/                   # html templates
├── swagger.yaml                 # API documentation
├── makefile                     # Dev task shortcuts
├── Dockerfile, docker-compose.yml
├── scripts/                     # Deploy scripts
├── cdk/                         # AWS infrastructure (Python CDK)
└── go.mod / go.sum              # Dependencies
```
