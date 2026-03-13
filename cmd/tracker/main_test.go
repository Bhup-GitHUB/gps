package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gps/internal/location"

	"github.com/redis/go-redis/v9"
)

func TestStoreLocationWritesRedis(t *testing.T) {
	var key string
	var ttl time.Duration
	var body string

	app := App{
		setLocation: func(ctx context.Context, inputKey string, inputBody []byte, inputTTL time.Duration) error {
			key = inputKey
			ttl = inputTTL
			body = string(inputBody)
			return nil
		},
	}

	payload := []byte(`{"order_id":"order_1","rider_id":"rider_1","lat":12.9716,"lng":77.5946,"timestamp":1710000000}`)
	err := app.storeLocation(context.Background(), payload)
	if err != nil {
		t.Fatalf("store failed: %v", err)
	}

	if key != location.RedisKey("order_1") {
		t.Fatalf("expected rider key got %q", key)
	}

	if body == "" {
		t.Fatal("expected body to be stored")
	}

	if ttl != 30*time.Second {
		t.Fatalf("expected ttl up to 30s got %v", ttl)
	}
}

func TestHandleHealth(t *testing.T) {
	app := App{}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	res := httptest.NewRecorder()

	app.handleHealth(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", res.Code)
	}
}

func TestHandleLocationMissingOrderID(t *testing.T) {
	app := App{
		getLocation: func(ctx context.Context, key string) ([]byte, error) {
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/location", nil)
	res := httptest.NewRecorder()

	app.handleLocation(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", res.Code)
	}
}

func TestHandleLocationNotFound(t *testing.T) {
	app := App{
		getLocation: func(ctx context.Context, key string) ([]byte, error) {
			return nil, redis.Nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/location?order_id=order_404", nil)
	res := httptest.NewRecorder()

	app.handleLocation(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", res.Code)
	}
}

func TestHandleLocationFound(t *testing.T) {
	app := App{
		getLocation: func(ctx context.Context, key string) ([]byte, error) {
			return []byte(`{"order_id":"order_1","rider_id":"rider_1","lat":12.9716,"lng":77.5946,"timestamp":1710000000}`), nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/location?order_id=order_1", nil)
	res := httptest.NewRecorder()

	app.handleLocation(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", res.Code)
	}
}

func TestHandleLocationReadFailed(t *testing.T) {
	app := App{
		getLocation: func(ctx context.Context, key string) ([]byte, error) {
			return nil, errors.New("read failed")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/location?order_id=order_1", nil)
	res := httptest.NewRecorder()

	app.handleLocation(res, req)

	if res.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 got %d", res.Code)
	}
}
