package location

import (
	"testing"
	"time"
)

func TestEventValid(t *testing.T) {
	lat := 12.9716
	lng := 77.5946

	event := Event{
		OrderID: "order_1",
		RiderID: "rider_1",
		Lat:     &lat,
		Lng:     &lng,
	}

	if !event.Valid() {
		t.Fatal("expected event to be valid")
	}
}

func TestEventNormalize(t *testing.T) {
	lat := 12.9716
	lng := 77.5946

	event := Event{
		OrderID: "order_1",
		RiderID: "rider_1",
		Lat:     &lat,
		Lng:     &lng,
	}

	event.Normalize()

	if event.Timestamp == 0 {
		t.Fatal("expected timestamp to be set")
	}
}

func TestEncodeDecode(t *testing.T) {
	lat := 12.9716
	lng := 77.5946

	event := Event{
		OrderID:   "order_1",
		RiderID:   "rider_1",
		Lat:       &lat,
		Lng:       &lng,
		Timestamp: time.Now().Unix(),
	}

	body, err := Encode(event)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	decoded, err := Decode(body)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded.OrderID != event.OrderID {
		t.Fatalf("expected %q got %q", event.OrderID, decoded.OrderID)
	}

	if decoded.RiderID != event.RiderID {
		t.Fatalf("expected %q got %q", event.RiderID, decoded.RiderID)
	}
}

func TestRedisKey(t *testing.T) {
	key := RedisKey("order_1")
	if key != "rider:order_1:loc" {
		t.Fatalf("expected rider:order_1:loc got %q", key)
	}
}
