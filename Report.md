# Отчёт

## 1. Общая информация

В рамках данной задачи я запустил своё приложение в однодовой инсталляции Kubernetes - minikube.

Все манифесты находятся в директории k8s/

Основные компоненты приложения:
- frontend;
- backend;
- PostgreSQL;
- миграции базы данных.

---

## 2. Какие абстракции Kubernetes использованы для запуска

### 2.1 Deployment

Deployment используется для:

- PostgreSQL;
- backend;
- frontend.

Он управляет подами и раскатывает наше приложение с желаемым количеством реплик и желаемыми параметрами.


Для backend количество реплик может быть изменено командой:

```bash
kubectl scale deployment backend --replicas=3
```

После масштабирования нам добавляются ещё 2 пода, приложение продолжает корректно отрабатывать.

---

### 2.2 Service

Для сетевого взаимодействия между компонентами используются Service.

Созданы следующие Service:

- db
- backend
- frontend


`db` и `backend` имеют тип ClusterIP, чтобы внутри самого кластера уметь обращаться друг к другу по сети, так как поды эфемерные и айпишники могут меняться

`frontend` имеет тип NodePort - чтобы еще отдельно аллоцировать порт на ноде по которому мы можем подключаться к приложению с хоста

Service выбирает Pod по селектору и лэблам.
---

### 2.3 PersistentVolumeClaim

Для PostgreSQL используется PersistentVolumeClaim(PVC): postgres-data

Объём хранилища: 1Gi
Режим доступа: ReadWriteOnce


---

### 2.4 ConfigMap

ConfigMap используется для хранения переменных окружения, например:
```text
DB_HOST
DB_PORT
DB_USER
DB_NAME
HTTP_PORT
```

Контейнер backend получает значения оттуда.

---

### 2.5 Secret

Secret используется для хранения DB_PASSWORD и  JWT_SECRET.

---

### 2.6 Job

Для выполнения миграций базы данных используется Job:db-migrate - она создает под, который выполняет миграции и после этого остается Сompleted, а сама джоба Complete.

---

### 2.7 Readiness и liveness probes

Backend предоставляет:

```text
/health/live
/health/ready
```

Readiness probe используется для того, чтобы проверять готов ли бэк обслуживать запросы.

Liveness probe используется для того, чтобы понять а жив ли бэкенд процесс.


---

## 3. Взаимодействие компонентов

Frontend:

Пользователь открывает приложение через Service `frontend` типа `NodePort`. 

nginx:

nginx обслуживает статические файлы frontend и проксирует запросы к API с префиксом `/api` на внутренний Kubernetes Service `backend` по адресу `backend:8080`, то еесть нет зависимости от конкретного IP пода. 

Backend:

Backend получает адрес бд из `DB_HOST=db` и подключается к PostgreSQL по адресу `db:5432`. Имя `db` разрешается внутри кластера в Service PostgreSQL, который направляет соединение в текущий готовый PostgreSQL Pod.

PostgreSQL:

PostgreSQL запущен в отдельном Pod под управлением Deployment. Для доступа к нему используется Service `db` типа `ClusterIP`, поэтому база доступна только внутри Kubernetes-кластера.

Миграции:

Миграции базы данных выполняются отдельным Kubernetes Job `db-migrate`.

---

## 4. Масштабирование backend

Backend масштабируется с одной до трёх реплик:

```bash
kubectl scale deployment backend --replicas=3
```

---

## 5. Запуск приложения

Манифесты применяются командой:

```bash
kubectl apply -f k8s/
```

Проверка соответствующих абстракций:

```bash
kubectl get pods

kubectl get deployments

kubectl get svc

kubectl get pvc

kubectl get jobs
```
ы
Для открытия frontend используется:

```bash
minikube service frontend
```

---

## 6. Результат

В результате:

- Приложение работает
- backend масштабируется
- данные сохраняются в PV, и не зависят от удаления подов