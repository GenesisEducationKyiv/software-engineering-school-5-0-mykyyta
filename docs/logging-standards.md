# Microservices Logging Standards

## Overview
This document defines the logging standards for all microservices in the weather API system, based on the implementation patterns established in the email service.

## Core Principles

### 1. **Structured Logging**
- Use structured logging with key-value pairs, not string interpolation
- All logs must be machine-readable JSON format in production
- Use consistent field names across all services

### 2. **Context-Based Logging**
- Always use context to pass logger instances through the call chain
- Use `logger.With(ctx, logger)` and `logger.From(ctx)` pattern
- Enrich logger with service-specific context (request IDs, user IDs, etc.)

### 3. **Privacy-First Approach**
- Never log sensitive data in plain text (emails, passwords, tokens, etc.)
- Use the `logger.HashEmail()` function for email addresses
- Sanitize any PII before logging

## Log Levels Standards

### **DEBUG Level**
Use for detailed tracing that's only needed during development or troubleshooting:
- Normal HTTP requests (successful ones)
- Database query execution details
- External API calls (successful ones)
- Message processing steps
- Template rendering details
- Business logic flow tracing

```go
logger.From(ctx).Debugw("Processing user request", 
    "user", logger.HashEmail(userEmail), 
    "action", "weather_fetch")
```

### **INFO Level**
Use only for significant application events:
- Service startup/shutdown
- Consumer/worker start/stop
- Configuration loading
- Health check status changes
- Scheduled job execution

```go
logger.From(ctx).Infow("Service started", 
    "service", "weather", 
    "port", cfg.Port)
```

### **WARN Level**
Use for recoverable issues that need attention:
- Client errors (4xx HTTP responses)
- Slow operations (>1s for HTTP, >5s for background tasks)
- Circuit breaker activation
- Retry attempts
- Deprecated API usage
- Resource limits approaching

```go
logger.From(ctx).Warnw("Slow HTTP request", 
    "method", r.Method, 
    "path", r.URL.Path, 
    "duration_ms", duration.Milliseconds())
```

### **ERROR Level**
Use for actual failures requiring immediate attention:
- Server errors (5xx HTTP responses)
- External service failures
- Database connection errors
- Message processing failures
- Template rendering errors
- Unhandled exceptions

```go
logger.From(ctx).Errorw("External API failed", 
    "service", "weather_provider", 
    "err", err,
    "retry_count", retryCount)
```

## Middleware Logging Standards

### HTTP Request Middleware
```go
func RequestMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Generate/extract request ID
        reqID := r.Header.Get("X-Request-ID")
        if reqID == "" {
            reqID = uuid.NewString()
        }

        // Enrich logger with request context
        logger := loggerPkg.From(r.Context()).With("request_id", reqID)
        ctx := loggerPkg.With(r.Context(), logger)
        w.Header().Set("X-Request-ID", reqID)

        start := time.Now()
        ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
        next.ServeHTTP(ww, r.WithContext(ctx))
        duration := time.Since(start)

        // Log based on response status and duration
        if ww.status >= 500 {
            logger.Errorw("HTTP request failed", 
                "method", r.Method, "path", r.URL.Path, 
                "status", ww.status, "duration_ms", duration.Milliseconds())
        } else if ww.status >= 400 {
            logger.Warnw("HTTP client error", 
                "method", r.Method, "path", r.URL.Path, 
                "status", ww.status, "duration_ms", duration.Milliseconds())
        } else if duration > 1000*time.Millisecond {
            logger.Warnw("Slow HTTP request", 
                "method", r.Method, "path", r.URL.Path, 
                "status", ww.status, "duration_ms", duration.Milliseconds())
        } else {
            logger.Debugw("HTTP request", 
                "method", r.Method, "path", r.URL.Path, 
                "status", ww.status, "duration_ms", duration.Milliseconds())
        }
    })
}
```

## Handler Logging Standards

### Request Processing
```go
func (h *Handler) ProcessRequest(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    logger := loggerPkg.From(ctx)

    var req RequestType
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logger.Warnw("Invalid JSON in request body", "err", err)
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }

    // Validate request
    if err := validateRequest(req); err != nil {
        logger.Warnw("Invalid request", "validation_error", err)
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    // Process request (errors logged in service layer)
    if err := h.service.Process(ctx, req); err != nil {
        logger.Errorw("Request processing failed", "err", err)
        http.Error(w, "processing failed", http.StatusInternalServerError)
        return
    }

    // Success - log at DEBUG level only
    logger.Debugw("Request processed successfully")
    w.WriteHeader(http.StatusOK)
}
```

## Service Layer Logging Standards

### Business Logic
```go
func (s *Service) ProcessBusinessLogic(ctx context.Context, req Request) error {
    // Debug level for normal flow
    logger.From(ctx).Debugw("Starting business process", 
        "process_type", req.Type,
        "user", logger.HashEmail(req.UserEmail))

    // Call external services
    data, err := s.externalClient.Fetch(ctx, req.Params)
    if err != nil {
        logger.From(ctx).Errorw("External service failed", 
            "service", "external_api", 
            "err", err)
        return err
    }

    // Process data
    result, err := s.processor.Process(ctx, data)
    if err != nil {
        logger.From(ctx).Errorw("Data processing failed", "err", err)
        return err
    }

    // Success - debug level
    logger.From(ctx).Debugw("Business process completed", 
        "result_count", len(result))
    return nil
}
```

## Consumer/Worker Logging Standards

### Message Processing
```go
func (c *Consumer) processMessage(ctx context.Context, msg Message) error {
    // Add message ID to context
    logger := loggerPkg.From(ctx).With("message_id", msg.ID)
    ctx = loggerPkg.With(ctx, logger)

    // Debug level for normal message receipt
    logger.Debugw("Message received", 
        "topic", msg.Topic, 
        "size", len(msg.Body))

    start := time.Now()
    err := c.processBusinessLogic(ctx, msg)
    duration := time.Since(start)

    if err != nil {
        logger.Errorw("Message processing failed", 
            "err", err, 
            "duration_ms", duration.Milliseconds())
        return err
    }

    // Log slow processing as warning
    if duration > 5*time.Second {
        logger.Warnw("Slow message processing", 
            "duration_ms", duration.Milliseconds())
    } else {
        logger.Debugw("Message processed successfully", 
            "duration_ms", duration.Milliseconds())
    }

    return nil
}
```

## External Adapter Logging Standards

### API Clients
```go
func (c *APIClient) CallExternalService(ctx context.Context, req Request) (*Response, error) {
    // Debug level for normal API calls
    logger.From(ctx).Debugw("Calling external API", 
        "endpoint", c.baseURL+req.Path,
        "method", req.Method)

    resp, err := c.httpClient.Do(req.HTTPRequest)
    if err != nil {
        logger.From(ctx).Errorw("External API request failed", 
            "endpoint", c.baseURL+req.Path,
            "err", err)
        return nil, err
    }

    if resp.StatusCode >= 400 {
        logger.From(ctx).Errorw("External API error response", 
            "endpoint", c.baseURL+req.Path,
            "status_code", resp.StatusCode)
        return nil, fmt.Errorf("API error: %d", resp.StatusCode)
    }

    // Success - debug level
    logger.From(ctx).Debugw("External API call successful", 
        "endpoint", c.baseURL+req.Path,
        "status_code", resp.StatusCode)
    
    return parseResponse(resp), nil
}
```

## Database Logging Standards

### Repository Layer
```go
func (r *Repository) FindUser(ctx context.Context, email string) (*User, error) {
    // Debug level for normal queries
    logger.From(ctx).Debugw("Executing database query", 
        "table", "users", 
        "operation", "find_by_email",
        "user", logger.HashEmail(email))

    var user User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            // Not found is not an error, just debug info
            logger.From(ctx).Debugw("User not found", 
                "user", logger.HashEmail(email))
            return nil, nil
        }
        
        logger.From(ctx).Errorw("Database query failed", 
            "table", "users", 
            "operation", "find_by_email",
            "err", err)
        return nil, err
    }

    logger.From(ctx).Debugw("Database query successful", 
        "table", "users", 
        "operation", "find_by_email")
    
    return &user, nil
}
```

## Privacy and Security Standards

### Sensitive Data Handling
```go
// ✅ DO: Hash sensitive data
logger.Infow("User registered", "user", logger.HashEmail(email))

// ❌ DON'T: Log sensitive data in plain text
logger.Infow("User registered", "email", email)

// ✅ DO: Sanitize before logging
logger.Debugw("Processing payment", "card_last4", card[len(card)-4:])

// ❌ DON'T: Log full sensitive data
logger.Debugw("Processing payment", "card_number", cardNumber)
```

### Error Context
```go
// ✅ DO: Include relevant context without sensitive data
logger.Errorw("Authentication failed", 
    "user", logger.HashEmail(email),
    "attempt_count", attempts,
    "ip", clientIP)

// ❌ DON'T: Log passwords or tokens
logger.Errorw("Authentication failed", 
    "email", email,
    "password", password) // Never do this!
```

## Configuration Standards

### Logger Initialization
```go
func main() {
    env := os.Getenv("ENV")
    logger, err := loggerPkg.New(loggerPkg.Config{
        Service: "service-name", // Use consistent service names
        Env:     env,
        Level:   os.Getenv("LOG_LEVEL"), // Allow runtime log level control
    })
    if err != nil {
        panic(err)
    }
    defer logger.Sync()

    logger.Infow("Service starting", 
        "service", "service-name",
        "env", env,
        "version", buildVersion)

    // Pass logger to application
    if err := app.Run(logger); err != nil {
        logger.Fatalw("Service crashed", "err", err)
    }
}
```

## Field Naming Conventions

### Standard Field Names
- `service` - Service name (e.g., "email", "weather", "subscription")
- `request_id` - HTTP request correlation ID
- `message_id` - Message queue correlation ID
- `user` - Hashed user identifier (use `logger.HashEmail()`)
- `err` - Error messages
- `duration_ms` - Operation duration in milliseconds
- `status_code` - HTTP status codes
- `method` - HTTP methods
- `path` - HTTP paths
- `table` - Database table names
- `operation` - Operation type (e.g., "find", "create", "update")

### Performance Fields
- `duration_ms` - Always in milliseconds
- `count` - Item counts
- `size` - Data sizes in bytes
- `retry_count` - Retry attempt numbers

## Testing and Monitoring

### Health Checks
```go
func (s *Service) HealthCheck(ctx context.Context) error {
    // Only log health check failures, not successes
    if err := s.db.Ping(); err != nil {
        logger.From(ctx).Errorw("Health check failed", 
            "component", "database", 
            "err", err)
        return err
    }
    
    // Don't log successful health checks - they create noise
    return nil
}
```

### Metrics Integration
```go
// Log metrics-worthy events at INFO level
logger.From(ctx).Infow("Email sent", 
    "template", templateName,
    "provider", "sendgrid") // This can feed metrics systems
```

## Implementation Checklist

For each new service, ensure:

- [ ] Request middleware with request ID generation
- [ ] Appropriate log levels (DEBUG for normal ops, ERROR for failures)
- [ ] Sensitive data sanitization using `logger.HashEmail()`
- [ ] Consistent field naming conventions
- [ ] Context-based logger propagation
- [ ] Performance logging (duration, slow operation warnings)
- [ ] External service failure logging
- [ ] Database operation logging
- [ ] Graceful service startup/shutdown logging

## Common Antipatterns to Avoid

❌ **DON'T:**
- Log every successful operation at INFO level
- Log sensitive data in plain text
- Use string concatenation instead of structured fields
- Log without context
- Create custom log levels
- Log stack traces for expected errors
- Log the same information at multiple levels

✅ **DO:**
- Use DEBUG for normal operations
- Hash sensitive data before logging
- Use structured key-value pairs
- Pass logger through context
- Use standard log levels appropriately
- Include relevant context in error logs
- Log once at the appropriate level

This standard ensures consistent, privacy-safe, and operationally useful logging across all microservices.
