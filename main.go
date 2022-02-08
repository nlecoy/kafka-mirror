package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/segmentio/kafka-go"
)

// Config provides the system configuration.
type Config struct {
	Destination struct {
		Brokers string `envconfig:"KAFKA_MIRROR_DESTINATION_BROKERS"`
		Topic   string `envconfig:"KAFKA_MIRROR_DESTINATION_TOPIC" required:"false"`
	}
	SampleRate float64 `envconfig:"KAFKA_MIRROR_SAMPLE_RATE" default:"1.0"`
	Source     struct {
		Brokers string `envconfig:"KAFKA_MIRROR_SOURCE_BROKERS"`
		Topic   string `envconfig:"KAFKA_MIRROR_SOURCE_TOPIC"`
	}
}

func sample(rate float64) bool {
	return rand.Float64() < rate
}

func main() {
	var envfile string
	flag.StringVar(&envfile, "env-file", ".env", "Read in a file of environment variables.")
	flag.Parse()

	godotenv.Load(envfile)
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	rand.Seed(time.Now().UnixNano())
	rate := cfg.SampleRate
	if rate < 0 || rate > 1.0 {
		rate = 1.0
	}

	destinationTopic := cfg.Destination.Topic
	if destinationTopic == "" {
		destinationTopic = cfg.Source.Topic
	}

	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  strings.Split(cfg.Source.Brokers, ","),
		Topic:    cfg.Source.Topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer consumer.Close()

	producer := &kafka.Writer{
		Addr:         kafka.TCP(strings.Split(cfg.Destination.Brokers, ",")...),
		Topic:        destinationTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 10 * time.Millisecond,
	}
	defer producer.Close()

	fmt.Println("start consuming...")
	for {
		msg, err := consumer.ReadMessage(ctx)
		if err != nil {
			log.Fatal(err)
		}

		if !sample(rate) {
			continue
		}

		err = producer.WriteMessages(ctx, kafka.Message{
			Key:   msg.Key,
			Value: msg.Value,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}
