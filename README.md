# Weather API - Microservices Architecture

A weather subscription service built with Go that allows users to subscribe to weather updates via email. The system is designed with a microservices architecture using Docker containers.

## 🏗️ Architecture

This project implements a **microservices architecture** with the following services:

- **Gateway Service** (`gateway/`) - API Gateway and routing
- **Weather Service** (`weather/`) - Weather data retrieval via external APIs  
- **Subscription Service** (`subscription/`) - User subscription management
- **Email Service** (`email/`) - Email notifications via SendGrid

### Tech Stack

- **Language**: Go 1.24.3
- **Framework**: Gin (HTTP routing)
- **Database**: PostgreSQL (subscriptions), Redis (caching)
- **Message Queue**: RabbitMQ
- **Communication**: gRPC (inter-service)
- **Containerization**: Docker & Docker Compose
- **Monitoring**: Prometheus, Grafana, Loki

## ✨ Features

- 📧 **Email Subscriptions** - Users can subscribe with email and city
- ✅ **Double Opt-in** - Secure email confirmation via JWT tokens
- 🌤️ **Weather Updates** - Daily/hourly weather notifications
- 🔄 **Unsubscribe** - Easy unsubscription via secure links
- 📊 **Monitoring** - Comprehensive logging and metrics
- 🚀 **Scalable** - Microservices architecture for horizontal scaling

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.24+ (for development)
- Make (optional, for convenience commands)

### Environment Setup

1. **Clone the repository**
```bash
git clone <repository-url>
cd weatherApi
```

2. **Setup environment variables**
Create `.env` files in each service directory:
- `microservices/weather/.env`
- `microservices/subscription/.env` 
- `microservices/email/.env`

**Run instruction** 

make up # Start all services in the background

make down # Stop and remove containers

make restart # Restart services with rebuild

make logs # View logs

make build # Build images without starting

## 🛠️ Development

### Project Structure
```
microservices/
├── gateway/          # API Gateway
├── weather/          # Weather data service  
├── subscription/     # Subscription management
├── email/           # Email notifications
├── proto/           # gRPC protocol definitions
└── pkg/             # Shared packages (logger, metrics)
```

## 🔧 Configuration

### Docker Services
- **PostgreSQL** (Port 5432) - Main database
- **Redis** (Port 6379) - Caching layer  
- **RabbitMQ** (Port 5672, Management: 15672) - Message queue
- **Prometheus** (Port 9090) - Metrics collection
- **Grafana** (Port 3000) - Metrics visualization

### Monitoring & Observability
- Structured logging with Zap
- Prometheus metrics collection
- Grafana dashboards
- Health checks for all services

## 📋 Project Status

This project demonstrates:
- ✅ Microservices architecture
- ✅ gRPC inter-service communication  
- ✅ Docker containerization
- ✅ Database migrations
- ✅ Comprehensive testing
- ✅ Monitoring and observability
- 🚧 Full documentation (in progress)

**Note:** This project is still a work in progress. Full documentation will be added later.