# System Design — Weather API

## 1. Overview

This document explains how the Weather API system works. It’s a backend-only service that allows users to subscribe for weather updates via email. Users can confirm their email using a secure link and receive weather updates for a specific city, either hourly or daily.

---

## 2. Requirements

Detailed requirements are stored in separate files:

- [Functional Requirements](./requirements/functional.md)
- [Non-Functional Requirements](./requirements/non-functional.md)

---

## 3. Architecture and Modules

The application is organized into modules. Each folder has its own responsibility.

| Module / Path            | Description                                                        |
|--------------------------|--------------------------------------------------------------------|
| `cmd/server/`            | Entry point — starts the HTTP server                               |
| `config/`                | Loads environment settings                                         |
| `internal/api/`          | HTTP routes: subscribe, confirm, unsubscribe, weather              |
| `internal/model/`        | Database models: Subscription, Weather (GORM)                      |
| `internal/db/`           | Database connection setup                                          |
| `pkg/email/`             | Handles sending emails via SendGrid                                |
| `pkg/jwtutil/`           | JWT token generation and validation                                |
| `pkg/scheduler/`         | Periodic task for sending emails                                   |
| `pkg/weatherapi/`        | Weather data provider (real or mock)                               |
| `templates/`             | Static HTML form (`subscribe.html`) for user email subscriptions   |
| `scripts/`               | Deployment helpers (e.g. upload secrets, redeploy)                 |
| `cdk/`                   | Infrastructure as Code (AWS CDK in Python)                         |

---

## 4. System Behavior

[View Flow Sequence Diagram](./diagrams/flow-sequence-diagram.svg)

### Subscription Flow

1. The user fills out `subscribe.html` with their email and city.
2. The backend saves the subscription and sets `IsConfirmed = false`.
3. It generates a secure JWT token and stores it in the DB.
4. An email with the confirmation link is sent using SendGrid.
5. When the user clicks the link, the backend marks the subscription as confirmed.

---

### Email Delivery (Scheduled)

1. A background task runs every hour or day.
2. It finds active subscriptions (`IsConfirmed` + not unsubscribed).
3. For each:
    - It fetches weather data.
    - Composes a personalized email.
    - Sends the email via SendGrid.

---

### Unsubscribe

1. Each email has an unsubscribe link with a token.
2. When the user clicks it, the backend:
    - Checks the token,
    - Marks the user as unsubscribed.

---

### Token Handling

- Tokens are JWTs, stored in the database.
- They don’t expire.
- When a request comes in with a token:
    - The backend decodes it,
    - Looks up the subscription by ID,
    - Compares the token with the one saved in the DB.

---

## 5. Data Models

### Subscription

Stores user data and subscription preferences.

| Field           | Description                                      |
|------------------|--------------------------------------------------|
| `ID`             | UUID string (unique)                             |
| `Email`          | User’s email, must be unique                     |
| `City`           | Chosen city                                      |
| `Frequency`      | "daily" or "hourly"                              |
| `IsConfirmed`    | Whether the user confirmed their email           |
| `IsUnsubscribed` | Whether the user opted out                       |
| `Token`          | JWT stored in DB (not shown in responses)        |
| `CreatedAt`      | When the subscription was created                |

### Weather

Holds the weather data returned via API or email.

| Field         | Description                     |
|----------------|---------------------------------|
| `Temperature` | Temperature in Celsius           |
| `Humidity`    | Humidity (0–100%)                |
| `Description` | Weather summary (e.g. "Clear")   |

---

## 6. Authentication

- Only used for confirmation and unsubscribe via JWT tokens.
- No login or sessions.
- Tokens are checked by comparing against saved DB value.

---

## 7. Deployment

### Local

- Run with Docker Compose (`docker-compose.yml`)
- Uses `.env` file for secrets and config

### Production (AWS)

- Infrastructure is defined in Python CDK (`cdk/`)
- Environment variables are uploaded to AWS SSM
- Two helper scripts:
    - `upload_env_to_ssm.sh`: uploads local `.env` to AWS
    - `redeploy_ecs.sh`: restarts the ECS service

---

## 8. Limitations / TODO

- No retries for failed email sends
- Old or unconfirmed subscriptions not automatically removed

---

## 9. References

- [ADR Index](../adr/000-title-index.md)
- [API Reference — Swagger YAML](../swagger.yaml)
- [Infrastructure (CDK)](../cdk/)