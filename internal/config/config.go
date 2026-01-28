package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Env             string
	PostgresDSN     string
	KafkaBrokers    []string
	KafkaTopic      string
	KafkaGroupID    string
	GRPCPort        int
	ConsumerWorkers int
}

func Load() (Config, error) {
	// Best-effort load for local development. In CI/tests you often want hermetic env;
	// set HOOKIFY_SKIP_DOTENV=1 to disable reading .env.
	if strings.TrimSpace(os.Getenv("HOOKIFY_SKIP_DOTENV")) != "1" {
		_ = godotenv.Load()
	}

	env := strings.TrimSpace(os.Getenv("HOOKIFY_ENV"))
	if env == "" {
		env = "development"
	}

	postgresDSN := strings.TrimSpace(os.Getenv("HOOKIFY_POSTGRES_DSN"))
	if postgresDSN == "" {
		return Config{}, errors.New("HOOKIFY_POSTGRES_DSN is required")
	}

	brokersRaw := strings.TrimSpace(os.Getenv("HOOKIFY_KAFKA_BROKERS"))
	if brokersRaw == "" {
		return Config{}, errors.New("HOOKIFY_KAFKA_BROKERS is required")
	}

	var brokers []string
	for _, part := range strings.Split(brokersRaw, ",") {
		b := strings.TrimSpace(part)
		if b != "" {
			brokers = append(brokers, b)
		}
	}
	if len(brokers) == 0 {
		return Config{}, errors.New("HOOKIFY_KAFKA_BROKERS must contain at least one broker")
	}

	topic := strings.TrimSpace(os.Getenv("HOOKIFY_KAFKA_TOPIC"))
	if topic == "" {
		return Config{}, errors.New("HOOKIFY_KAFKA_TOPIC is required")
	}

	groupID := strings.TrimSpace(os.Getenv("HOOKIFY_KAFKA_GROUP_ID"))
	if groupID == "" {
		return Config{}, errors.New("HOOKIFY_KAFKA_GROUP_ID is required")
	}

	grpcPort := 50051
	if v := strings.TrimSpace(os.Getenv("HOOKIFY_GRPC_PORT")); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid HOOKIFY_GRPC_PORT: %w", err)
		}
		grpcPort = p
	}

	workers := 1
	if v := strings.TrimSpace(os.Getenv("HOOKIFY_CONSUMER_WORKERS")); v != "" {
		w, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid HOOKIFY_CONSUMER_WORKERS: %w", err)
		}
		if w <= 0 {
			return Config{}, errors.New("HOOKIFY_CONSUMER_WORKERS must be > 0")
		}
		workers = w
	}

	return Config{
		Env:             env,
		PostgresDSN:     postgresDSN,
		KafkaBrokers:    brokers,
		KafkaTopic:      topic,
		KafkaGroupID:    groupID,
		GRPCPort:        grpcPort,
		ConsumerWorkers: workers,
	}, nil
}
