package setup

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	postgresImage    = "postgres:15"
	postgresDatabase = "db_name"
	postgresUsername = "username"
	postgresPassword = "password"
	postgresPort     = "5432"
	postgresHost     = "localhost"
)

type PostgresConfig struct {
	User        string
	Password    string
	Host        string
	ExposedPort string
	LocalPort   string
	Name        string
}

func SetupPostgres(ctx context.Context) (_ PostgresConfig, _ func(ctx context.Context) error, err error) {
	container, err := postgres.Run(
		ctx,
		postgresImage,
		postgres.WithDatabase(postgresDatabase),
		postgres.WithUsername(postgresUsername),
		postgres.WithPassword(postgresPassword),
		testcontainers.WithWaitStrategyAndDeadline(
			time.Minute,
			wait.ForLog("database system is ready to accept connections"),
		),
	)
	if err != nil {
		return PostgresConfig{}, nil, fmt.Errorf("run postgres container: %w", err)
	}

	shutdown := func(ctx context.Context) error {
		if err := container.Terminate(ctx); err != nil {
			return fmt.Errorf("terminate postgres container: %w", err)
		}
		return nil
	}

	defer func() {
		if err == nil {
			return
		}
		if errShutdown := shutdown(ctx); errShutdown != nil {
			err = errors.Join(err, errShutdown)
		}
	}()

	port, err := container.MappedPort(ctx, postgresPort+"/tcp")
	if err != nil {
		return PostgresConfig{}, nil, fmt.Errorf("get postgres exposed port: %w", err)
	}

	config := PostgresConfig{
		User:        postgresUsername,
		Password:    postgresPassword,
		Host:        postgresHost,
		ExposedPort: port.Port(),
		LocalPort:   postgresPort,
		Name:        postgresDatabase,
	}

	return config, shutdown, nil
}
