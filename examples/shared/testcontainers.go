package shared

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const defaultServerPort = "6379/tcp"

// RedisImage returns the Redis image used by the kvx examples.
func RedisImage() string {
	return envOrDefault("KVX_REDIS_IMAGE", "redis:7-alpine")
}

// RedisJSONImage returns the Redis Stack image used by the kvx JSON examples.
func RedisJSONImage() string {
	return envOrDefault("KVX_REDIS_JSON_IMAGE", "redis/redis-stack-server:latest")
}

// ValkeyImage returns the Valkey image used by the kvx examples.
func ValkeyImage() string {
	return envOrDefault("KVX_VALKEY_IMAGE", "valkey/valkey:8-alpine")
}

// ValkeyJSONImage returns the Valkey image used by the kvx JSON examples.
func ValkeyJSONImage() string {
	return envOrDefault("KVX_VALKEY_JSON_IMAGE", "valkey/valkey:8-alpine")
}

// StartContainer starts a kvx example container and returns its address.
func StartContainer(ctx context.Context, image string) (testcontainers.Container, string, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        image,
			ExposedPorts: []string{defaultServerPort},
			WaitingFor:   wait.ForListeningPort(defaultServerPort).WithStartupTimeout(45 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		return nil, "", fmt.Errorf("start %s container: %w", image, err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, "", terminateContainerWithError(
			ctx,
			container,
			fmt.Errorf("resolve %s container host: %w", image, err),
		)
	}

	port, err := container.MappedPort(ctx, defaultServerPort)
	if err != nil {
		return nil, "", terminateContainerWithError(
			ctx,
			container,
			fmt.Errorf("resolve %s container port: %w", image, err),
		)
	}

	return container, fmt.Sprintf("%s:%s", host, port.Port()), nil
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func terminateContainerWithError(
	ctx context.Context,
	container testcontainers.Container,
	err error,
) error {
	terminateErr := container.Terminate(ctx)
	if terminateErr == nil {
		return err
	}

	return errors.Join(err, fmt.Errorf("terminate container: %w", terminateErr))
}
