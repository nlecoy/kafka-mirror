package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
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

func main() {
	var envfile string
	flag.StringVar(&envfile, "env-file", ".env", "Read in a file of environment variables.")
	flag.Parse()

	godotenv.Load(envfile)
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	sourceTopic := cfg.Source.Topic
	if sourceTopic == "" {
		log.Fatal("must provide a source topic")
	}

	destinationTopic := cfg.Destination.Topic
	if destinationTopic == "" {
		destinationTopic = sourceTopic
	}

	fmt.Println("Source topic", sourceTopic)
	fmt.Println("Destination topic", destinationTopic)

	consumer := kafka.NewReader(kafka.ReaderConfig{
		Brokers: strings.Split(cfg.Source.Brokers, ","),
		Topic:   sourceTopic,
		GroupID: "kafka-mirror",
	})
	defer consumer.Close()

	consumer.SetOffset(0)

	producer := &kafka.Writer{
		Addr:         kafka.TCP(strings.Split(cfg.Destination.Brokers, ",")...),
		Topic:        destinationTopic,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: time.Second,
	}
	defer producer.Close()

	fmt.Println("Start consuming...")
	for {
		msg, err := consumer.FetchMessage(ctx)
		if err != nil {
			log.Fatal(err)
		}

		err = producer.WriteMessages(ctx, kafka.Message{
			Key:   msg.Key,
			Value: msg.Value,
		})
		if err != nil {
			log.Fatalf("Error writing to Kafka: %s", err)
		}
	}
}
