package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"gps/internal/location"

	"github.com/segmentio/kafka-go"
)

type App struct {
	writer *kafka.Writer
	topic  string
}

func main() {
	port := os.Getenv("INGEST_PORT")
	if port == "" {
		port = "8080"
	}

	broker := os.Getenv("KAFKA_BROKERS")
	if broker == "" {
		broker = "localhost:9092"
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "rider_locations"
	}

	writer := &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.Hash{},
	}

	app := App{
		writer: writer,
		topic:  topic,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/location", app.handleLocation)

	log.Println("ingest api on", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func (a App) handleLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	log.Println("location hit")

	var event location.Event

	err := json.NewDecoder(r.Body).Decode(&event)
	if err != nil {
		log.Println("invalid payload")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	event.Normalize()

	if !event.Valid() {
		log.Println("missing fields")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing fields"})
		return
	}

	body, err := json.Marshal(event)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to encode"})
		return
	}

	err = a.writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(event.OrderID),
		Value: body,
	})
	if err != nil {
		log.Println("produce failed", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to queue"})
		return
	}

	log.Println("produced event", event.OrderID, a.topic)
	writeJSON(w, http.StatusAccepted, map[string]any{"ok": true, "status": "queued"})
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(body)
}
