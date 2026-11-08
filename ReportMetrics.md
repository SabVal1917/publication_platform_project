# Отчёт: метрики приложения

## Стек

Go + chi, PostgreSQL 15, frontend на Nginx, кластер minikube.
VictoriaMetrics Single установлена через Helm в namespace `monitoring`:
чарт `victoria-metrics-single` 0.49.0, приложение v1.154.0.
Хранение: PVC 2 GiB, срок - 7 дней.

## Экспорт и сбор

Backend использует `prometheus/client_golang v1.23.2`.
GoCollector и ProcessCollector экспортируют метрики рантайма и процесса.

Ручка `/metrics` поддерживает Prometheus text и согласование OpenMetrics.

Она доступна без JWT через внутренний ClusterIP; frontend проксирует `/api/`.

PostgreSQL exporter `quay.io/prometheuscommunity/postgres-exporter:v0.19.0`
работает вторым контейнером Pod БД: подключается к localhost:5432/blog,
публикует метрики на порту 9187, получает пароль из Kubernetes Secret.

Встроенный сборщик VictoriaMetrics опрашивает таргеты каждые 15 секунд.

Service discovery обнаруживает Pod по аннотациям `prometheus.io/scrape`,
`prometheus.io/port`, `prometheus.io/path`; порты backend/db - 8080/9187.

## Бизнес-метрики

Все метрики ниже имеют префикс `publication_platform_` и тип Counter.

| Имя | Когда увеличивается |
| --- | --- |
| `users_registered_total` | Учётная запись сохранена в БД |
| `logins_total` | Проверен пароль и сформирован JWT при входе |
| `articles_created_total` | Транзакция создания статьи завершена успешно |
| `comments_created_total` | Транзакция создания комментария завершена успешно |

Ошибки не увеличивают соответствующий счётчик; перезапуск сбрасывает его.
Существовавшие до внедрения метрик записи не учитываются.

## Развёртывание и проверка

Манифесты приложения и exporter: `k8s/default/`.
Helm values и снимок ресурсов VictoriaMetrics: `k8s/monitoring/`.

```bash
helm upgrade --install vm-single vm/victoria-metrics-single --version 0.49.0 --namespace monitoring --create-namespace -f k8s/monitoring/victoria-metrics-values.yaml
kubectl get pods,svc,pvc -n monitoring
kubectl port-forward -n monitoring svc/victoria-metrics 8428:8428
```

VMUI: http://localhost:8428/vmui/; статус сбора: http://localhost:8428/targets.

Метрики проверил.

Все тесты прошли.