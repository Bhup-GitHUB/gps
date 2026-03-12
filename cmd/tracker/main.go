package main

import (
	"context"
	"encoding/json"
	"log"
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

	log.Println("consumer started", topic)
	app.runConsumer(context.Background())
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
