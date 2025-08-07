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