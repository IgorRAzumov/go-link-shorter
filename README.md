# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Swagger

Документация API доступна по адресу `/swagger/index.html` при запущенном сервере.

Регенерация документации после изменения аннотаций в хендлерах:

```bash
go generate ./internal/controller/rest/...
```

Или вручную:

```bash
swag init -g internal/controller/rest/docs.go -o docs --parseDependency --parseInternal
```

## Профилирование памяти (pprof)

1. Снять профиль до оптимизации:

```bash
go test -bench=BenchmarkAPIHandler_Shorten -memprofile=profiles/base.pprof -memprofilerate=1 -run=NONE ./internal/controller/rest/
```

2. Изучить профиль (top, list, web, peek):

```bash
go tool pprof -top profiles/base.pprof
go tool pprof -list=. profiles/base.pprof
go tool pprof -web profiles/base.pprof
go tool pprof -peek=. profiles/base.pprof
```

3. Снять профиль после оптимизации:

```bash
go test -bench=BenchmarkAPIHandler_Shorten -memprofile=profiles/result.pprof -memprofilerate=1 -run=NONE ./internal/controller/rest/
```

4. Проверка результата оптимизации:

```bash
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
```

**Вывод команды pprof -top -diff_base=profiles/base.pprof profiles/result.pprof** (отрицательные значения — снижение
потребления памяти):

```
File: rest.test
Type: alloc_space
Time: 2026-02-19 20:51:42 EET
Showing nodes accounting for 26.76MB, 21.68% of 123.43MB total
Dropped 78 nodes (cum <= 0.62MB)
      flat  flat%   sum%        cum   cum%
   10.59MB  8.58%  8.58%    10.59MB  8.58%  bufio.NewReaderSize (inline)
    3.23MB  2.62% 11.20%     3.23MB  2.62%  net/http.(*Request).WithContext (partial-inline)
    1.29MB  1.05% 12.25%     1.29MB  1.05%  encoding/json.(*Decoder).refill
    1.09MB  0.88% 13.13%     1.09MB  0.88%  net/http.Header.Clone (inline)
    1.09MB  0.88% 14.02%     1.09MB  0.88%  net/url.parse
    0.97MB  0.79% 14.80%     0.97MB  0.79%  net/textproto.MIMEHeader.Set (inline)
    0.93MB  0.75% 15.55%     0.93MB  0.75%  net/textproto.MIMEHeader.Add (inline)
    0.85MB  0.69% 16.24%     0.85MB  0.69%  strings.(*Builder).grow
    0.81MB  0.65% 16.90%     0.81MB  0.65%  encoding/json.NewDecoder
    0.81MB  0.65% 17.55%     1.39MB  1.12%  net/http.readRequest
    0.65MB  0.52% 18.08%     0.65MB  0.52%  crypto/internal/fips140/sha256.New (inline)
    0.57MB  0.46% 18.53%     1.21MB  0.98%  crypto/internal/fips140/hmac.New
    0.40MB  0.33% 18.86%     0.40MB  0.33%  net/http/httptest.NewRecorder (inline)
    0.36MB  0.29% 19.16%     0.36MB  0.29%  context.WithValue
    0.32MB  0.26% 19.42%     0.32MB  0.26%  encoding/hex.EncodeToString (inline)
    0.32MB  0.26% 19.68%     2.34MB  1.90%  auth.(*Service).SignUserID
    0.32MB  0.26% 19.94%     0.32MB  0.26%  github.com/mailru/easyjson/buffer.getBuf
    0.28MB  0.23% 20.17%     0.28MB  0.23%  encoding/base64.(*Encoding).EncodeToString
    0.24MB   0.2% 20.37%     1.74MB  1.41%  shorter.GenerateShortKey
    0.24MB   0.2% 20.57%     0.24MB   0.2%  github.com/google/uuid.UUID.String (inline)
    0.24MB   0.2% 20.76%     0.24MB   0.2%  internal/sync.newEntryNode
    0.20MB  0.16% 20.93%     0.20MB  0.16%  adapter.FromDomainLink (inline)
    0.16MB  0.13% 21.06%     0.16MB  0.13%  bytes.(*Buffer).grow
    0.16MB  0.13% 21.19%     1.78MB  1.44%  auth.(*Service).hmacHex
    0.16MB  0.13% 21.32%     1.01MB  0.82%  shorter.(*Service).CreateShortKey
    0.12MB 0.098% 21.42%     0.16MB  0.13%  shorter.(*Service).GenerateShortKey
    0.08MB 0.065% 21.48%     0.69MB  0.56%  inmemory.(*LinkStorage).Save
    0.08MB 0.065% 21.55%    11.37MB  9.22%  middleware.HTTPLogger.func1
    0.08MB 0.065% 21.61%    12.99MB 10.52%  net/http/httptest.NewRequestWithContext
    0.04MB 0.033% 21.65%     6.27MB  5.08%  benchRouter.APIHandler.func2
    0.04MB 0.033% 21.68%     0.28MB  0.23%  authctx.WithUserID (partial-inline)
         0     0% 21.68%    -0.18MB  0.15%  github.com/google/uuid.New (inline)
         0     0% 21.68%     1.04MB  0.84%  github.com/google/uuid.NewString
         0     0% 21.68%    27.69MB 22.43%  BenchmarkAPIHandler_Shorten
         0     0% 21.68%    27.69MB 22.43%  testing.(*B).runN
```

## Профилирование БД (pprof)

Требуется запущенный PostgreSQL (по умолчанию: postgres, пароль пустой, порт 5432). Скрипт откатывает оптимизации,
снимает base, восстанавливает код и снимает result:

```bash
./scripts/profile_db.sh
```

Или вручную:

1. Снять base:
   `go test -bench=BenchmarkBatchSave -memprofile=profiles/db_base.pprof -memprofilerate=1 -run=NONE ./internal/adapter/database/ -benchtime=2s`
2. Снять result: то же с `profiles/db_result.pprof`
3. Сравнение: `go tool pprof -top -diff_base=profiles/db_base.pprof profiles/db_result.pprof`

**Вывод pprof -top -diff_base для БД** (отрицательные значения — снижение потребления памяти):

```
File: database.test
Type: alloc_space
      flat  flat%   sum%        cum   cum%
   -2850kB  8.31% 125.61% 47002.11kB 136.97%  database.(*Storage).BatchSave
   -2850kB  8.31% 117.30% -11753.39kB 34.25%  github.com/lib/pq.(*stmt).ExecContext
-2110.45kB  6.15% 117.53% -2110.45kB  6.15%  strings.genSplit
  -1425kB  4.15% 124.20% -20303.58kB 59.17%  database/sql.resultFromStatement
  -712.50kB  2.08%  database/sql.(*Tx).grabConn
  -698.67kB  2.04% 140.71%  -698.67kB  2.04%  github.com/lib/pq.(*readBuf).string
  -712.62kB  2.08%  github.com/google/uuid.New (inline)
-20988.70kB 61.17%  database/sql.(*DB).retry
-21016.08kB 61.25%  database/sql.(*Stmt).ExecContext
```

Бенчмарк: base 357 iter, 2172 allocs/op → result 488 iter, 1873 allocs/op. Меньше аллокаций (−299), выше throughput (
+37%).

## Анализ результатов оптимизации

**BenchmarkAPIHandler_Shorten (память):** По diff pprof видно снижение аллокаций в `uuid.New` (−0.18MB). Основные
затраты остаются в HTTP-стеке (bufio, Request, json.Decoder), middleware и auth — это штатная работа обработчика.

**BenchmarkBatchSave (БД):** Оптимизация — bulk INSERT вместо построчных вставок. Один запрос
`INSERT ... VALUES ($1..$4), ($5..$8), ...` вместо множества `ExecContext` по строкам. Снижение: ~21MB по
`database/sql` (подготовка/выполнение statement), меньше `strings.genSplit` при сборке запроса. Итог: throughput +37% (
357→488 iter/s), аллокаций −299 на операцию (2172→1873).

