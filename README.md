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

## How to Run Locally

### 1️⃣ Setup environment

Copy the example file and review environment variables:

```bash
cp .env.example .env
```

There are two groups of variables:

**✅ Pre-filled defaults (should work for local testing):**

```env
PORT=8080  
DB_TYPE=postgres  
DB_URL=host=db user=postgres password=postgres dbname=weatherdb port=5432 sslmode=disable  
BASE_URL=http://localhost:8080  
GIN_MODE=debug  
JWT_SECRET=default_secret  
```

**❗ Required for full functionality (email, weather API):**

```env
SENDGRID_API_KEY=your_sendgrid_api_key_here  
EMAIL_FROM=no-reply@example.com  
WEATHER_API_KEY=your_weather_api_key_here  
```

> ℹ️ You can start the server without these keys, but email confirmation and weather data will not work until you provide them.

---

### 2️⃣ Install Go dependencies

```bash
go mod tidy
```

---

### 3️⃣ Run the project with Docker

```bash
docker-compose up --build
```
