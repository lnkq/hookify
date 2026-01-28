package config

import "testing"

func setBaseEnv(t *testing.T) {
	t.Helper()
	t.Setenv("HOOKIFY_SKIP_DOTENV", "1")
	t.Setenv("HOOKIFY_POSTGRES_DSN", "postgres://u:p@localhost:5432/db?sslmode=disable")
	t.Setenv("HOOKIFY_KAFKA_BROKERS", "k1:9092, k2:9092")
	t.Setenv("HOOKIFY_KAFKA_TOPIC", "topic")
	t.Setenv("HOOKIFY_KAFKA_GROUP_ID", "group")
	t.Setenv("HOOKIFY_GRPC_PORT", "50051")
}

func TestLoad_OK(t *testing.T) {
	setBaseEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if cfg.Env != "development" {
		t.Fatalf("expected default env=development, got %q", cfg.Env)
	}
	if cfg.GRPCPort != 50051 {
		t.Fatalf("expected GRPCPort=50051, got %d", cfg.GRPCPort)
	}
	if cfg.ConsumerWorkers != 1 {
		t.Fatalf("expected ConsumerWorkers=1, got %d", cfg.ConsumerWorkers)
	}
	if len(cfg.KafkaBrokers) != 2 {
		t.Fatalf("expected 2 kafka brokers, got %d", len(cfg.KafkaBrokers))
	}
}

func TestLoad_EnvProduction(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("HOOKIFY_ENV", "production")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if cfg.Env != "production" {
		t.Fatalf("expected env=production, got %q", cfg.Env)
	}
}

func TestLoad_MissingPostgresDSN(t *testing.T) {
	t.Setenv("HOOKIFY_SKIP_DOTENV", "1")
	t.Setenv("HOOKIFY_KAFKA_BROKERS", "k1:9092")
	t.Setenv("HOOKIFY_KAFKA_TOPIC", "topic")
	t.Setenv("HOOKIFY_KAFKA_GROUP_ID", "group")

	_, err := Load()
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestLoad_InvalidWorkers(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("HOOKIFY_CONSUMER_WORKERS", "0")

	_, err := Load()
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestLoad_BrokersParsingTrims(t *testing.T) {
	setBaseEnv(t)
	t.Setenv("HOOKIFY_KAFKA_BROKERS", " a:1, ,b:2,  ")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(cfg.KafkaBrokers) != 2 || cfg.KafkaBrokers[0] != "a:1" || cfg.KafkaBrokers[1] != "b:2" {
		t.Fatalf("unexpected brokers: %#v", cfg.KafkaBrokers)
	}
}
