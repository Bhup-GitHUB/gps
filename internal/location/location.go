package location

import "time"

type Event struct {
	OrderID   string   `json:"order_id"`
	RiderID   string   `json:"rider_id"`
	Lat       *float64 `json:"lat"`
	Lng       *float64 `json:"lng"`
	Timestamp int64    `json:"timestamp"`
}

func (e *Event) Normalize() {
	if e.Timestamp == 0 {
		e.Timestamp = time.Now().Unix()
	}
}

func (e Event) Valid() bool {
	return e.OrderID != "" && e.RiderID != "" && e.Lat != nil && e.Lng != nil
}
