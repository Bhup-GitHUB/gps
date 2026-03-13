package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/segmentio/kafka-go"
)

func TestHandleLocationAccepted(t *testing.T) {
	var sent kafka.Message

	app := App{
		topic: "rider_locations",
		send: func(ctx context.Context, msg kafka.Message) error {
			sent = msg
			return nil
		},
	}

	body := []byte(`{"order_id":"order_1","rider_id":"rider_1","lat":12.9716,"lng":77.5946,"timestamp":1710000000}`)
	req := httptest.NewRequest(http.MethodPost, "/location", bytes.NewReader(body))
	res := httptest.NewRecorder()

	app.handleLocation(res, req)

	if res.Code != http.StatusAccepted {
		t.Fatalf("expected 202 got %d", res.Code)
	}

	if string(sent.Key) != "order_1" {
		t.Fatalf("expected order_1 got %q", string(sent.Key))
	}

	var response map[string]any
	err := json.Unmarshal(res.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if response["status"] != "queued" {
		t.Fatalf("expected queued got %v", response["status"])
	}
}

func TestHandleLocationInvalidJSON(t *testing.T) {
	app := App{
		send: func(ctx context.Context, msg kafka.Message) error {
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/location", bytes.NewReader([]byte(`{"order_id":`)))
	res := httptest.NewRecorder()

	app.handleLocation(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", res.Code)
	}
}

func TestHandleLocationMissingFields(t *testing.T) {
	app := App{
		send: func(ctx context.Context, msg kafka.Message) error {
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/location", bytes.NewReader([]byte(`{"order_id":"order_1"}`)))
	res := httptest.NewRecorder()

	app.handleLocation(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", res.Code)
	}
}
