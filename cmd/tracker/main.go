package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"gps/internal/location"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type App struct {
	reader *kafka.Reader
	redis  *redis.Client
}

func main() {
	broker := os.Getenv("KAFKA_BROKERS")
	if broker == "" {
		broker = "localhost:9092"
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "rider_locations"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisDB := 0
	if os.Getenv("REDIS_DB") != "" {
		value, err := strconv.Atoi(os.Getenv("REDIS_DB"))
		if err == nil {
			redisDB = value
		}
	}

	port := os.Getenv("TRACKER_PORT")
	if port == "" {
		port = "8081"
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker},
		Topic:   topic,
		GroupID: "tracker-group",
	})

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       redisDB,
	})

	app := App{
		reader: reader,
		redis:  rdb,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", app.handleHealth)
	mux.HandleFunc("/location", app.handleLocation)

	log.Println("consumer started", topic)
	go app.runConsumer(context.Background())

	log.Println("tracker api on", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func (a App) runConsumer(ctx context.Context) {
	for {
		msg, err := a.reader.ReadMessage(ctx)
		if err != nil {
			log.Println("read failed", err)
			time.Sleep(time.Second)
			continue
		}

		log.Println("message received", string(msg.Key))

		err = a.storeLocation(ctx, msg.Value)
		if err != nil {
			log.Println("redis update failed", err)
			continue
		}

		log.Println("redis updated", string(msg.Key))
	}
}

func (a App) storeLocation(ctx context.Context, body []byte) error {
	var event location.Event

	err := json.Unmarshal(body, &event)
	if err != nil {
		return err
	}

	if !event.Valid() {
		return nil
	}

	return a.redis.Set(ctx, location.RedisKey(event.OrderID), body, 30*time.Second).Err()
}

func (a App) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a App) handleLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	orderID := r.URL.Query().Get("order_id")
	if orderID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order_id is required"})
		return
	}

	body, err := a.redis.Get(r.Context(), location.RedisKey(orderID)).Bytes()
	if err == redis.Nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "location not found"})
		return
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(body)
}
