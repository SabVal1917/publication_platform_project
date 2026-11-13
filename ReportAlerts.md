# Алерты publication-platform

## Что реализовано

Существующая VictoriaMetrics собирает метрики приложения и пробера каждые 15 секунд.
Добавлены vmalert для вычисления правил и Alertmanager для группировки и доставки
уведомлений в локальный webhook, включая сообщения о восстановлении `resolved`.
Webhook работает внутри кластера и записывает полученные JSON-уведомления в stdout.
Внешние аккаунты, токены и Secret для доставки не нужны.

Исходники приложения и Dockerfile не изменяются. Для демонстрации 500 и задержек
используется уже существующий образ `blog-backend:sre-demo` из предыдущей работы.

| Алерт | Уровень | Условие | Почему |
| --- | --- | --- | --- |
| `BackendHigh5xxCount` | critical | Не менее 5 ответов 5xx за последние 5 минут, условие держится 1 минуту | Серия серверных ошибок мешает работе пользователей; абсолютный порог подходит малонагруженному учебному сервису |
| `ProberUnavailable` | critical | Неуспешная проверка API, недоступный scrape пробера, отсутствие его метрик либо последняя проба старше 30 секунд; условие держится 1 минуту | Обнаруживает как недоступность приложения и чтения из БД, так и отказ самого пробера |
| `BackendHighLatency` | warning | p95 успешных ответов конкретного API-маршрута за 5 минут выше 500 мс в течение 2 минут; минимум 20 ответов за окно | Соответствует границе быстрых ответов из предыдущей работы; проверка числа запросов уменьшает шум при редких обращениях |
| `BackendGoHeapHigh` | warning | `go_memstats_heap_alloc_bytes` конкретного Pod выше 64 MiB в течение 2 минут | Алерт рантайма Go; порог выбран как половина текущего лимита контейнера 128 MiB, чтобы оставить запас для остальной памяти процесса |

Два алерта на метрики сервиса: `BackendHigh5xxCount`, `BackendHighLatency`.
Один алерт на метрики рантайма: `BackendGoHeapHigh` из GoCollector.
Пробер — дополнительный контроль доступности.
В формулировке задания названы Spring/Python, а в этом приложении используется Go;
GoCollector выполняет аналогичную задачу экспорта метрик рантайма.

5xx берутся из собственного Counter приложения
`publication_platform_http_requests_total{status=~"5.."}`.
`increase(...[5m])` считает прирост за окно с учётом сброса счётчика при рестарте,
а не общее количество ошибок за всё время. Результат может быть дробным из-за
расчёта прироста по scrape-данным. Никакие отдельные nginx-метрики не требуются.

Задержка вычисляется через `histogram_quantile(0.95, ...)` по существующим
`publication_platform_http_request_duration_seconds_bucket`. Гистограмма уже
содержит границу 0.5 секунды. Учитываются успешные ответы 2xx/3xx, отдельно по
маршруту, чтобы быстрые ошибки или другие ручки не скрывали медленный маршрут.
HTTP-правила ограничены `/api/.*` и исключают `/api/ping`, healthchecks и `/metrics`.

Для пробера проверяются `up`, `publication_platform_probe_success` и
`publication_platform_probe_timestamp_seconds`. `up` показывает успешность
сбора `/metrics`, а `probe_success` — результат HTTP-проверки `/api/articles`.
Подстановка `or vector(0)` обеспечивает срабатывание при исчезновении серии,
в том числе после масштабирования пробера до нуля.

Go heap — объём занятых объектов кучи, не весь RSS процесса и не потребление
памяти контейнером. При изменении лимита 128 MiB порог нужно пересмотреть.
Порог 64 MiB — начальная настройка для этого стенда; далее его уточняют по
нормальному профилю нагрузки.

## Файлы

- `k8s/monitoring/alerts/rules/service.yaml` — три правила сервиса и пробера.
- `k8s/monitoring/alerts/rules/runtime.yaml` — правило рантайма Go.
- `k8s/monitoring/alerts/alertmanager-config.yaml` — маршрутизация на локальный webhook.
- `k8s/monitoring/alerts/alertmanager.yaml` — Deployment, Service и PVC Alertmanager.
- `k8s/monitoring/alerts/webhook.yaml` — Deployment и Service приёмника уведомлений.
- `k8s/monitoring/alerts/vmalert.yaml` — Deployment и Service vmalert.
- `k8s/monitoring/alerts/kustomization.yaml` — сборка ресурсов и ConfigMap.
- `k8s/monitoring/alerts/tests/alerts.test.yaml` — 11 сценариев проверки правил.
- `k8s/monitoring/alerts-memory-demo/` — временный тест памяти с порогом 1 байт.

Alertmanager группирует сообщения по имени алерта, namespace, сервису и уровню.
Первая отправка — через 10 секунд, изменения группы — через 30 секунд,
повтор неустранённого инцидента — раз в час. PVC сохраняет историю отправок
и silences при перезапуске. Inhibition не используется: для проверки каждого
условия нужны отдельные уведомления.

Kustomize добавляет hash конфигурации к именам ConfigMap. При изменении правил
и повторном `apply -k` vmalert автоматически перезапускается с новой конфигурацией.
Аналогично обновляется Alertmanager при изменении его конфига.
Существующие Helm-ресурсы VictoriaMetrics этими манифестами не заменяются.

## Подготовка стенда

Все команды выполняются из корня проекта. При проверке конфигурации minikube
был остановлен; следующие шаги запускает пользователь. Срабатывания в работающем
кластере нужно подтвердить скриншотами. Доставка Alertmanager → webhook дополнительно
проверена локально в Docker; это не заменяет демонстрацию алертов в minikube.

```bash
minikube start
kubectl get pods -A
kubectl get svc victoria-metrics -n monitoring
```

Если VictoriaMetrics ещё не установлена, использовать команду из `ReportMetrics.md`.
В существующем scrape-конфиге `kubernetes-pods` должны сохраняться labels
`namespace`, `app`, `pod`: именно так настроен сохранённый в репозитории
`victoria-metrics-rendered.yaml`.

Развернуть Alertmanager, vmalert и приёмник webhook:

```bash
kubectl apply -k k8s/monitoring/alerts
kubectl rollout status deployment/alert-webhook -n monitoring --timeout=180s
kubectl rollout status deployment/alertmanager -n monitoring --timeout=180s
kubectl rollout status deployment/vmalert -n monitoring --timeout=180s
```

Убедиться, что готовый демо-образ присутствует, загрузить его в minikube и
переключить backend на него. Это загрузка готового образа, без сборки.
Если backend уже использует `blog-backend:sre-demo`, переключение не требуется.

```bash
docker image inspect blog-backend:sre-demo --format '{{.Id}}'
minikube image load blog-backend:sre-demo
kubectl set image deployment/backend backend=blog-backend:sre-demo -n default
kubectl rollout status deployment/backend -n default --timeout=180s
kubectl apply -f k8s/sre/prober.yaml
kubectl rollout status deployment/sre-prober -n default --timeout=180s
```

Открыть четыре отдельных терминала с port-forward и оставить их работающими:

```bash
kubectl port-forward -n monitoring svc/victoria-metrics 8428:8428
```

```bash
kubectl port-forward -n monitoring svc/vmalert 8880:8880
```

```bash
kubectl port-forward -n monitoring svc/alertmanager 9093:9093
```

```bash
kubectl port-forward -n default svc/backend 8080:8080
```

Интерфейсы:

- `http://localhost:8880/vmalert/groups` — правила, выражения и состояния.
- `http://localhost:9093/#/alerts` — активные алерты в Alertmanager.
- `http://localhost:8428/vmui/` — запросы и графики метрик.
- `http://localhost:8428/targets` — сбор метрик backend и пробера.

В отдельном терминале открыть логи получателя уведомлений:

```bash
kubectl logs -n monitoring deployment/alert-webhook --follow --since=10m
```

Для читаемого скриншота с установленным `jq` использовать:

```bash
kubectl logs -n monitoring deployment/alert-webhook --follow --since=10m | jq --unbuffered -R 'fromjson? | select(.json.alerts != null) | .json | {status, receiver, commonLabels, alerts: [.alerts[] | {status, labels, annotations, startsAt, endsAt}]}'
```

Фильтр скрывает стартовую строку сервера и оставляет тело уведомления Alertmanager.
Нужно увидеть `receiver: "local-webhook"`, `status: "firing"`, имя алерта,
`severity` и описание условия. После восстановления придёт `status: "resolved"`.
Без `jq` тело уведомления видно в сырой строке лога в поле `json`.
Не добавлять `--timestamps` при использовании этого фильтра: строка должна
оставаться JSON без префикса времени.

Логи хранятся в контейнере и могут быть потеряны при замене Pod или ротации.
Для сохранения свидетельства можно отдельно записать их в файл:

```bash
kubectl logs -n monitoring deployment/alert-webhook > /tmp/alert-notifications.log
```

В VMUI проверить, что следующие запросы возвращают серии:

```promql
up{job="kubernetes-pods",namespace="default",app=~"backend|sre-prober"}
```

```promql
go_memstats_heap_alloc_bytes{job="kubernetes-pods",namespace="default",app="backend"}
```

Прогреть метрики нормальным трафиком:

```bash
bash scripts/sre-demo.sh normal 180
```

На странице правил четыре алерта должны быть `inactive`. У правила памяти
есть результат запроса к реальной метрике, даже когда алерт не активен.
Если интерфейс показывает ошибки вычисления, сначала проверить datasource,
labels и логи vmalert. Если пробер успел прислать стартовый алерт, дождаться
`resolved` перед основным сценарием.

## Проверка 5xx и неуспешной пробы

```bash
bash scripts/sre-demo.sh error 180
```

В течение примерно 1–2 минут `BackendHigh5xxCount` и `ProberUnavailable`
перейдут из `pending` в `firing`. Пробер при этом работает, но получает HTTP 500.
В логах webhook должны появиться два уведомления `firing` с уровнем `critical`.

Снять скриншоты, пока команда ещё работает:

1. Страница vmalert: раскрытые правила `BackendHigh5xxCount` и
   `ProberUnavailable`, выражения и состояние `firing`.
2. Alertmanager: оба активных алерта с labels `severity=critical`.
3. Терминал с логами webhook: уведомления `firing` с именами алертов и описанием условий.

Скрипт автоматически возвращает режим `normal`, в том числе при Ctrl+C.
Ошибка 5xx может оставаться активной ещё около 5 минут после окончания нагрузки,
пока ошибки не выйдут из скользящего окна. Это ожидаемое поведение.

```bash
bash scripts/sre-demo.sh normal 360
```

Дождаться `inactive` и сообщений `resolved`. Снять скрин логов webhook с `status: "resolved"`.
Это подтверждает полный цикл обнаружения и устранения инцидента.

## Проверка отказа самого пробера

При одной исходной реплике:

```bash
kubectl scale deployment/sre-prober -n default --replicas=0
```

Дождаться `ProberUnavailable` и уведомления webhook `firing`. После исчезновения Pod
серия может оставаться видимой некоторое время, поэтому проверке нужно дать
несколько минут. Ветка устаревшего timestamp обнаруживает остановившиеся пробы.
Снять скриншоты алерта и сообщения. Восстановить пробер:

```bash
kubectl scale deployment/sre-prober -n default --replicas=1
kubectl rollout status deployment/sre-prober -n default --timeout=180s
```

Дождаться `resolved`. Если исходных реплик было больше одной, вернуть исходное число.

## Проверка задержки

```bash
bash scripts/sre-demo.sh slow 360
```

Готовый демо-образ добавляет задержку 800 мс на `GET /api/articles`.
Это больше порога 500 мс, но меньше таймаута пробера 2 секунды.
`BackendHighLatency` должен перейти в `firing` примерно через 2–4 минуты.
p95 по гистограмме приблизительный и не обязан показывать ровно 0.800 с.
Если стенд перегружен, пробер также может сработать по таймауту.

Пока нагрузка работает, снять:

1. vmalert: выражение `BackendHighLatency`, `route=/api/articles`, `firing`.
2. Логи webhook: `status: "firing"`, имя алерта и уровень `warning`.
3. По желанию VMUI: график p95, используя часть выражения правила
   `histogram_quantile(...)` без сравнения и без проверки количества запросов.

После завершения дать окну обновиться:

```bash
bash scripts/sre-demo.sh normal 360
```

Дождаться `resolved` для задержки.

## Проверка алерта рантайма Go

Без изменений кода нет управляемого способа заставить Go heap занять 64 MiB.
Для проверки используется реальная метрика GoCollector с временным порогом
1 байт и `for: 30s`. Это тест правила и доставки уведомления, а не доказательство
реального превышения рабочего порога. Тест явно помечен `test=memory-demo`,
а описание уведомления содержит тестовый и рабочий пороги.

```bash
kubectl apply -k k8s/monitoring/alerts-memory-demo
kubectl rollout status deployment/vmalert -n monitoring --timeout=180s
```

После замены Pod перезапустить port-forward vmalert в его терминале:

```bash
kubectl port-forward -n monitoring svc/vmalert 8880:8880
```

Примерно через 1–2 минуты снять:

1. VMUI: результат `go_memstats_heap_alloc_bytes` для backend, с Pod и значением.
2. vmalert: `BackendGoHeapHigh`, выражение с `> 1`, состояние `firing`,
   labels `component=go-runtime`, `test=memory-demo`.
3. Логи webhook: `status: "firing"` для `BackendGoHeapHigh` с тестовым описанием и `warning`.

Вернуть рабочую конфигурацию:

```bash
kubectl apply -k k8s/monitoring/alerts
kubectl rollout status deployment/vmalert -n monitoring --timeout=180s
```

Ещё раз перезапустить port-forward vmalert. Снять правило памяти с рабочим
порогом 64 MiB и `for: 2m`. После рестарта тестовое уведомление может перейти
в `resolved` с задержкой до нескольких минут: старая тестовая группа больше
не обновляется и закрывается по времени окончания алерта.

Если для оценки потребуется именно превышение 64 MiB, этот тест нужно дополнить
реальной нагрузкой на данные или отдельным согласованным способом выделения
памяти в Go-процессе. Выделение памяти в стороннем контейнере не увеличит Go heap
backend и не проверит этот алерт.

## Что приложить к сдаче

1. Каталог `k8s/monitoring/alerts/` с правилами, конфигурацией Alertmanager
   и Kubernetes-манифестами получателя webhook.
2. Таблицу условий и обоснование из начала этого документа.
3. Скриншот списка четырёх созданных правил с рабочими порогами.
4. Скриншоты состояния `firing` и полученных webhook-уведомлений для 5xx, пробера и latency.
5. Скриншоты метрики Go, активного тестового правила памяти и его уведомления;
   подписать, что порог временно снижен для проверки.
6. Скриншот сообщения `resolved` после устранения одного из инцидентов.

Состояние `firing` в Alertmanager показывает приём алерта; скриншот логов
`alert-webhook` отдельно подтверждает доставку уведомления по HTTP. В подписи
указать: «Уведомление Alertmanager получено локальным webhook, показано тело JSON».

Если уведомлений нет, проверить Pod и Service webhook, endpoints и логи компонентов:

```bash
kubectl get pods,svc -n monitoring
kubectl get endpointslices -n monitoring -l kubernetes.io/service-name=alert-webhook
kubectl logs deployment/alertmanager -n monitoring --tail=100
kubectl logs deployment/vmalert -n monitoring --tail=100
kubectl logs deployment/alert-webhook -n monitoring --tail=100
```

После проверки можно оставить готовый демо-образ в режиме `normal`.
Если до проверки backend работал на `blog-backend:dev`, вернуть его:

```bash
kubectl set image deployment/backend backend=blog-backend:dev -n default
kubectl rollout status deployment/backend -n default --timeout=180s
```

## Выполненная проверка конфигураций

Обе сборки Kustomize успешно отрендерены; ссылки Deployment на сгенерированные
ConfigMap проверены. Официальный `vmalert:v1.154.0` принял правила в режиме
`-dryRun`, `amtool` из `alertmanager:v0.28.1` принял конфигурацию Alertmanager.
`vmalert-tool:v1.154.0` успешно выполнил 11 сценариев на изолированных данных:
исправный, неуспешный, недоступный и исчезнувший пробер, устаревшая проба,
превышение и граница памяти, 5xx, исключение ping/4xx, задержка и недостаток запросов.
В одноразовом Docker-стенде отдельно проверена доставка Alertmanager → webhook
с этой же конфигурацией receiver и образом получателя: в логах получены уведомления
`firing` и `resolved`. Для этой проверки в Alertmanager вручную отправлен тестовый
алерт `AlertWebhookDeliveryTest`; реальные сбои приложения в Docker не создавались.
Одноразовый стенд после проверки удалён. Работа алертов на метриках приложения
в minikube проверяется по сценарию выше.

Повторить проверки из корня проекта:

```bash
kubectl kustomize k8s/monitoring/alerts > /tmp/publication-alerts.yaml
kubectl kustomize k8s/monitoring/alerts-memory-demo > /tmp/publication-alerts-memory-demo.yaml
docker run --rm -v "$PWD/k8s/monitoring/alerts/rules:/rules:ro" victoriametrics/vmalert:v1.154.0 -rule='/rules/*.yaml' -dryRun
docker run --rm -v "$PWD/k8s/monitoring/alerts/alertmanager-config.yaml:/config/alertmanager.yaml:ro" --entrypoint /bin/amtool quay.io/prometheus/alertmanager:v0.28.1 check-config /config/alertmanager.yaml
docker run --rm -v "$PWD/k8s/monitoring/alerts:/alerts:ro" victoriametrics/vmalert-tool:v1.154.0 unittest --files=/alerts/tests/alerts.test.yaml
```

## Документация

- [vmalert: правила, datasource и отправка в Alertmanager](https://docs.victoriametrics.com/victoriametrics/vmalert/).
- [Alertmanager: маршрутизация и webhook receiver](https://prometheus.io/docs/alerting/latest/configuration/).
- [HTTP echo: запись тела HTTP-запроса в логи](https://github.com/mendhak/docker-http-https-echo).
- [GoCollector](https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/collectors).
