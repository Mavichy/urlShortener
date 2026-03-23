## URL Shortener
Сервис для сокращения ссылок.
### Запуск через консоль: 
- local memory
```bash
go run ./cmd/urlShortener -storage=memory -http-addr=:8080 -base-url=http://localhost:8080
```
- postgres
```bash
go run ./cmd/urlShortener -storage=postgres -postgres-dsn="postgres://postgres:postgres@localhost:5432/url_shortener?sslmode=disable" -http-addr=:8080 -base-url=http://localhost:8080```
```
### Запуск через Docker 
- local memory
```bash
docker build -t url-shortener .
docker run --rm -p 8080:8080 url-shortener -storage=memory -http-addr=:8080 -base-url=http://localhost:8080
```
- postgres
```bash
docker compose up --build 
```

### HTTP API:
  - POST /api/v1/links — создать короткую ссылку 

```json
{
  "url": "https://example.com/some/long/path"
}
```
  -  GET /{code} — получить оригинальный URL
```bash
curl http://localhost:8080/WhTIaxPDCw
```
