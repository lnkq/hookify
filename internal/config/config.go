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
	PostgresDSN  string
	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroupID string
	GRPCPort     int
}

func Load() (Config, error) {
	_ = godotenv.Load()

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

	return Config{
		PostgresDSN:  postgresDSN,
		KafkaBrokers: brokers,
		KafkaTopic:   topic,
		KafkaGroupID: groupID,
		GRPCPort:     grpcPort,
	}, nil
}
