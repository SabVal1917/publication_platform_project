# Отчёт: сбор логов в VictoriaLogs

## Реализация

Backend пишет JSON-логи в stdout с полями `timestamp`, `level`, `message`,
`service.name`, `request_id`, `method`, `route`, `status` и `duration_ms`.

VictoriaLogs развёрнута в single-node режиме с PVC. Vector запущен как
DaemonSet, читает логи backend из `/var/log`, добавляет Kubernetes-метаданные
и отправляет записи в VictoriaLogs. DaemonSet выбран, чтобы использовать один
сборщик на каждой ноде и не добавлять sidecar в под приложения.

## Основные команды

```bash
minikube image build -t blog-backend:dev ./backend
kubectl apply -k k8s/default
kubectl apply -k k8s/logging
kubectl get pods -A
kubectl port-forward -n observability svc/victoria-logs 9428:9428
```

Интерфейс: <http://localhost:9428/select/vmui/>

## LogsQL-запросы

### 1. Сырые логи за 15 минут

```logsql
_time:15m service.name:="publication-platform-backend"
```

Показывает все записи backend за последние 15 минут.

![Сырые логи](images/01-raw-logs.png)

### 2. Фильтр по статусу

```logsql
_time:15m service.name:="publication-platform-backend" status:="404"
```

Показывает запросы, завершившиеся ответом `404`.

![Логи со статусом 404](images/02-status-404.png)

### 3. Группировка по HTTP-статусу

```logsql
_time:15m service.name:="publication-platform-backend" status:*
| stats by (status) count() as requests
```

Считает количество запросов для каждого HTTP-статуса.

![Количество запросов по статусам](images/03-stats-by-status.png)

### 4. Квантили времени ответа

```logsql
_time:15m service.name:="publication-platform-backend" duration_ms:*
| stats
    quantile(0.5, duration_ms) as p50_ms,
    quantile(0.95, duration_ms) as p95_ms,
    quantile(0.99, duration_ms) as p99_ms
```

Показывает медиану, 95-й и 99-й перцентили времени ответа backend.

![Квантили времени ответа](images/04-duration-quantiles.png)
