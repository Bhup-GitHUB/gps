## GPS ZOMATO SWIGGGGGGGYYYY

run local

```bash
vercel dev
```

post location

```bash
curl -X POST http://localhost:3000/api/location \
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
curl "http://localhost:3000/api/location?order_id=order_1"
```
