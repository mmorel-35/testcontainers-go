package testcontainers_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/internal/core"
)

func TestGenericContainer_stop_start_withReuse(t *testing.T) {
	containerName := "my-nginx"

	opts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("8080/tcp"),
		testcontainers.WithReuseByName(containerName),
	}

	ctr, err := testcontainers.Run(t.Context(), nginxAlpineImage, opts...)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
	require.NotNil(t, ctr)

	err = ctr.Stop(t.Context(), nil)
	require.NoError(t, err)

	// Run another container with same container name:
	// The checks for the exposed ports must not fail when restarting the container.
	ctr1, err := testcontainers.Run(t.Context(), nginxAlpineImage, opts...)
	testcontainers.CleanupContainer(t, ctr1)
	require.NoError(t, err)
	require.NotNil(t, ctr1)
}

func TestGenericContainer_pause_start_withReuse(t *testing.T) {
	ctx := t.Context()
	containerName := "my-nginx"

	opts := []testcontainers.ContainerCustomizer{
		testcontainers.WithExposedPorts("8080/tcp"),
		testcontainers.WithReuseByName(containerName),
	}

	ctr, err := testcontainers.Run(ctx, nginxAlpineImage, opts...)
	testcontainers.CleanupContainer(t, ctr)
	require.NoError(t, err)
	require.NotNil(t, ctr)

	// Pause the container is not supported by our API, but we can do it manually
	// by using the Docker client.
	cli, err := core.NewClient(ctx)
	require.NoError(t, err)

	err = cli.ContainerPause(ctx, ctr.GetContainerID())
	require.NoError(t, err)

	// Because the container is paused, it should not be possible to start it again.
	ctr1, err := testcontainers.Run(ctx, nginxAlpineImage, opts...)
	testcontainers.CleanupContainer(t, ctr1)
	require.ErrorIs(t, err, errors.ErrUnsupported)
}
