package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gps/internal/config"
	"gps/internal/location"

	"github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
)

type App struct {
	reader      *kafka.Reader
	redis       *redis.Client
	setLocation func(context.Context, string, []byte, time.Duration) error
	getLocation func(context.Context, string) ([]byte, error)
}

func main() {
	broker := config.String("KAFKA_BROKERS", "localhost:9092")
	topic := config.String("KAFKA_TOPIC", "rider_locations")
	redisAddr := config.String("REDIS_ADDR", "localhost:6379")
	redisDB := config.Int("REDIS_DB", 0)
	port := config.String("TRACKER_PORT", "8081")

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
		setLocation: func(ctx context.Context, key string, body []byte, ttl time.Duration) error {
			return rdb.Set(ctx, key, body, ttl).Err()
		},
		getLocation: func(ctx context.Context, key string) ([]byte, error) {
			return rdb.Get(ctx, key).Bytes()
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", app.handleHealth)
	mux.HandleFunc("/location", app.handleLocation)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	log.Println("consumer started", topic)
	go func() {
		defer wg.Done()
		app.runConsumer(ctx)
	}()

	go func() {
		<-ctx.Done()
		log.Println("shutting tracker")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		server.Shutdown(shutdownCtx)
		app.reader.Close()
		app.redis.Close()
	}()

	log.Println("tracker api on", port)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	wg.Wait()
}

func (a App) runConsumer(ctx context.Context) {
	for {
		msg, err := a.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}

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
	event, err := location.Decode(body)
	if err != nil {
		return err
	}

	if !event.Valid() {
		return nil
	}

	return a.setLocation(ctx, location.RedisKey(event.OrderID), body, 30*time.Second)
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

	body, err := a.getLocation(r.Context(), location.RedisKey(orderID))
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
