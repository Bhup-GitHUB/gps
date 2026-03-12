package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"gps/internal/config"
	"gps/internal/location"

	"github.com/segmentio/kafka-go"
)

type App struct {
	writer *kafka.Writer
	topic  string
}

func main() {
	port := config.String("INGEST_PORT", "8080")
	broker := config.String("KAFKA_BROKERS", "localhost:9092")
	topic := config.String("KAFKA_TOPIC", "rider_locations")

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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		log.Println("shutting ingest")
		writer.Close()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	log.Println("ingest api on", port)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
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
