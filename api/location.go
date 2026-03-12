package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type LocationBody struct {
	OrderID   string   `json:"order_id"`
	RiderID   string   `json:"rider_id"`
	Lat       *float64 `json:"lat"`
	Lng       *float64 `json:"lng"`
	Timestamp int64    `json:"timestamp"`
}

type LocationData struct {
	OrderID   string  `json:"order_id"`
	RiderID   string  `json:"rider_id"`
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Timestamp int64   `json:"timestamp"`
}

var riderData = map[string]LocationData{}

func LocationHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		orderID := r.URL.Query().Get("order_id")
		if orderID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "order_id is required"})
			return
		}

		data, ok := riderData[orderID]
		if !ok {
			fmt.Println("rider miss", orderID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "location not found"})
			return
		}

		fmt.Println("rider hit", orderID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(data)
	case http.MethodPost:
		var body LocationBody

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
			return
		}

		if body.OrderID == "" || body.RiderID == "" || body.Lat == nil || body.Lng == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "missing fields"})
			return
		}

		if body.Timestamp == 0 {
			body.Timestamp = time.Now().Unix()
		}

		data := LocationData{
			OrderID:   body.OrderID,
			RiderID:   body.RiderID,
			Lat:       *body.Lat,
			Lng:       *body.Lng,
			Timestamp: body.Timestamp,
		}

		riderData[body.OrderID] = data

		fmt.Println("rider data", body.OrderID, body.RiderID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"ok": true, "order_id": body.OrderID})
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
	}
}
