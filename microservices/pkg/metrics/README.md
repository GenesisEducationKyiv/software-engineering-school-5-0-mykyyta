# Metrics Package

Пакет для збору Prometheus метрик HTTP запитів у мікросервісах.

## Основні можливості

- **HTTP Request metrics** - загальна кількість запитів з розбивкою по статусам
- **Request Duration** - гістограма тривалості запитів для аналізу продуктивності  
- **Error Tracking** - детальне відстеження помилок з типами
- **Active Connections** - моніторинг поточної кількості активних з'єднань

## Як використовувати

```go
import "your-project/pkg/metrics"

// Створити інстанс метрик
metrics := metrics.New(metrics.Config{
    Namespace: "weather_api",
    Subsystem: "http",
})

// У middleware
func RequestMiddleware(metrics *metrics.Metrics) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Відстеження активних коннекшенів
            metrics.IncActiveConnections("service-name", r.Method, r.URL.Path)
            defer metrics.DecActiveConnections("service-name", r.Method, r.URL.Path)
            
            start := time.Now()
            // ... обробка запиту ...
            duration := time.Since(start)
            
            // Запис метрик
            metrics.RecordRequest("service-name", r.Method, r.URL.Path, 
                fmt.Sprintf("%d", statusCode), duration)
            
            if statusCode >= 400 {
                metrics.RecordError("service-name", r.Method, r.URL.Path,
                    fmt.Sprintf("%d", statusCode), "client_error")
            }
        })
    }
}
```

## Доступні метрики

### 1. HTTP Requests Total
**Назва**: `{namespace}_{subsystem}_total`  
**Тип**: Counter  
**Мітки**: `service`, `method`, `path`, `status`  
**Призначення**: Загальна кількість HTTP запитів

### 2. Request Duration
**Назва**: `{namespace}_{subsystem}_duration_seconds`  
**Тип**: Histogram  
**Мітки**: `service`, `method`, `path`, `status`  
**Призначення**: Тривалість обробки запитів в секундах

### 3. Errors Total
**Назва**: `{namespace}_{subsystem}_errors_total`  
**Тип**: Counter  
**Мітки**: `service`, `method`, `path`, `status`, `error_type`  
**Призначення**: Кількість помилок з деталізацією по типам

### 4. Active Connections
**Назва**: `{namespace}_{subsystem}_active_connections`  
**Тип**: Gauge  
**Мітки**: `service`, `method`, `path`  
**Призначення**: Поточна кількість активних HTTP з'єднань

## Моніторинг та алерти

### Рекомендовані алерти

#### 🚨 Критичні (негайна реакція)

**1. Service Down**
```promql
# Відсутність запитів за останні 2 хвилини
absent_over_time(rate(http_requests_total[2m])) == 1
```
- **Причина**: Сервіс не обробляє запити або зупинився
- **Дія**: Негайна перевірка стану сервісу

**2. High Error Rate**
```promql
# Рівень помилок > 5% за останні 5 хвилин
(
  rate(http_requests_errors_total[5m]) / 
  rate(http_requests_total[5m])
) * 100 > 5
```
- **Причина**: Проблеми з БД, зовнішніми API або логікою
- **Дія**: Перевірка логів та стану залежностей

**3. High Response Time**
```promql
# P95 > 2 секунди за останні 5 хвилин
histogram_quantile(0.95, 
  rate(http_requests_duration_seconds_bucket[5m])
) > 2
```
- **Причина**: Перевантаження або проблеми продуктивності
- **Дія**: Аналіз навантаження та оптимізація

**4. Too Many Active Connections**
```promql
# Більше 100 активних з'єднань
http_requests_active_connections > 100
```
- **Причина**: Пікове навантаження або повільні запити
- **Дія**: Масштабування або оптимізація

#### ⚠️ Попереджувальні (моніторинг тенденцій)

**1. Slow Requests Increase**
```promql
# Більше 10 повільних запитів за хвилину (>1s)
increase(http_requests_duration_seconds_bucket{le="1.0"}[1m]) > 10
```

**2. Error Rate Trending Up**
```promql
# Зростання помилок за останню годину
increase(http_requests_errors_total[1h]) > 50
```

**3. High Memory/CPU Usage**
```promql
# Використання ресурсів > 80%
process_memory_usage_percent > 80
process_cpu_usage_percent > 80
```

### Grafana Dashboard Queries

#### Основні панелі

**Request Rate:**
```promql
rate(http_requests_total[5m])
```

**Error Rate:**
```promql
rate(http_requests_errors_total[5m]) / rate(http_requests_total[5m]) * 100
```

**Response Time Percentiles:**
```promql
histogram_quantile(0.50, rate(http_requests_duration_seconds_bucket[5m]))
histogram_quantile(0.95, rate(http_requests_duration_seconds_bucket[5m]))
histogram_quantile(0.99, rate(http_requests_duration_seconds_bucket[5m]))
```

**Active Connections:**
```promql
http_requests_active_connections
```

#### Детальний аналіз

**Top Slow Endpoints:**
```promql
topk(10, 
  histogram_quantile(0.95, 
    rate(http_requests_duration_seconds_bucket[5m])
  )
) by (path)
```

**Error Distribution by Status:**
```promql
rate(http_requests_errors_total[5m]) by (status)
```

## Рекомендовані пороги алертів

| Метрика | Warning | Critical |
|---------|---------|----------|
| Error Rate | > 2% | > 5% |
| Response Time P95 | > 1s | > 2s |
| Active Connections | > 50 | > 100 |
| Service Down | - | 2 minutes |

## Retention Policy

- **High-resolution metrics** (15s): 7 днів
- **Medium-resolution** (1m): 30 днів  
- **Low-resolution** (5m): 90 днів
- **Long-term** (1h): 1 рік

## Best Practices

1. **Labels Cardinality** - уникайте високої кардинальності міток (ID, timestamps)
2. **Path Normalization** - нормалізуйте шляхи для REST endpoints (`/user/:id` замість `/user/123`)
3. **Error Classification** - використовуйте `error_type` для категоризації помилок
4. **Resource Limits** - встановлюйте ліміти на кількість метрик для уникнення OOM

## Інтеграція з іншими системами

### Alertmanager
Метрики автоматично експортуються в Prometheus та можуть використовуватись в Alertmanager для сповіщень.

### Jaeger Tracing
Рекомендується поєднувати з distributed tracing для повного розуміння запитів.

### Logging
Корелюйте метрики з логами через `correlation_id` для швидшого debugging-у.
