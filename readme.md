## GPS ZOMATO SWIGGGGGGGYYYY

defaults are in `.env.example`

```bash
cat .env.example
```

start kafka and redis

```bash
docker-compose up -d
```

run tracker

```bash
go run ./cmd/tracker
```

run ingest

```bash
go run ./cmd/ingest
```

post location

```bash
curl -X POST http://localhost:8080/location \
  -H "Content-Type: application/json" \
  -d '{
    "order_id":"order_1",
    "rider_id":"rider_1",
    "lat":12.9716,
    "lng":77.5946,
    "timestamp":1710000000
  }'
```

get location

```bash
curl "http://localhost:8081/location?order_id=order_1"
```

health

```bash
curl "http://localhost:8081/health"
```
