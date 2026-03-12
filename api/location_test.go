package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func resetRiderData() {
	riderLock.Lock()
	riderData = map[string]LocationData{}
	riderLock.Unlock()
}

func TestLocationPostAndGet(t *testing.T) {
	resetRiderData()

	body := []byte(`{"order_id":"order_1","rider_id":"rider_1","lat":12.9716,"lng":77.5946,"timestamp":1710000000}`)
	postReq := httptest.NewRequest(http.MethodPost, "/api/location", bytes.NewReader(body))
	postRes := httptest.NewRecorder()

	LocationHandler(postRes, postReq)

	if postRes.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", postRes.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/location?order_id=order_1", nil)
	getRes := httptest.NewRecorder()

	LocationHandler(getRes, getReq)

	if getRes.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", getRes.Code)
	}

	var data LocationData
	err := json.Unmarshal(getRes.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if data.OrderID != "order_1" {
		t.Fatalf("expected order_1 got %q", data.OrderID)
	}

	if data.RiderID != "rider_1" {
		t.Fatalf("expected rider_1 got %q", data.RiderID)
	}
}

func TestLocationPostOverwrite(t *testing.T) {
	resetRiderData()

	first := []byte(`{"order_id":"order_1","rider_id":"rider_1","lat":12.9716,"lng":77.5946,"timestamp":1710000000}`)
	second := []byte(`{"order_id":"order_1","rider_id":"rider_1","lat":13.0011,"lng":77.7000,"timestamp":1710000010}`)

	firstReq := httptest.NewRequest(http.MethodPost, "/api/location", bytes.NewReader(first))
	firstRes := httptest.NewRecorder()
	LocationHandler(firstRes, firstReq)

	secondReq := httptest.NewRequest(http.MethodPost, "/api/location", bytes.NewReader(second))
	secondRes := httptest.NewRecorder()
	LocationHandler(secondRes, secondReq)

	getReq := httptest.NewRequest(http.MethodGet, "/api/location?order_id=order_1", nil)
	getRes := httptest.NewRecorder()
	LocationHandler(getRes, getReq)

	var data LocationData
	err := json.Unmarshal(getRes.Body.Bytes(), &data)
	if err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if data.Lat != 13.0011 {
		t.Fatalf("expected 13.0011 got %v", data.Lat)
	}
}

func TestLocationGetMissingOrderID(t *testing.T) {
	resetRiderData()

	req := httptest.NewRequest(http.MethodGet, "/api/location", nil)
	res := httptest.NewRecorder()

	LocationHandler(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", res.Code)
	}
}

func TestLocationGetNotFound(t *testing.T) {
	resetRiderData()

	req := httptest.NewRequest(http.MethodGet, "/api/location?order_id=order_404", nil)
	res := httptest.NewRecorder()

	LocationHandler(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected 404 got %d", res.Code)
	}
}

func TestLocationPostInvalidJSON(t *testing.T) {
	resetRiderData()

	req := httptest.NewRequest(http.MethodPost, "/api/location", bytes.NewReader([]byte(`{"order_id":`)))
	res := httptest.NewRecorder()

	LocationHandler(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", res.Code)
	}
}

func TestLocationPostMissingFields(t *testing.T) {
	resetRiderData()

	req := httptest.NewRequest(http.MethodPost, "/api/location", bytes.NewReader([]byte(`{"order_id":"order_1"}`)))
	res := httptest.NewRecorder()

	LocationHandler(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d", res.Code)
	}
}
