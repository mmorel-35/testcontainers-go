package testcontainers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
)

// ExampleCleanupContainer_recommended demonstrates the RECOMMENDED pattern for
// cleaning up containers in tests using CleanupContainer.
//
// This pattern ensures containers are properly terminated even if the test fails.
func ExampleCleanupContainer_recommended() {
	ctx := context.Background()

	// Create container
	container, err := testcontainers.Run(ctx,
		"nginx:alpine",
		testcontainers.WithExposedPorts("80/tcp"),
	)
	// IMPORTANT: Call CleanupContainer immediately after creation, before error check
	// This ensures cleanup happens even if err != nil
	testcontainers.CleanupContainer(&testing.T{}, container)
	if err != nil {
		panic(err)
	}

	// Use container in your test...
	// When test ends (success or failure), container will be automatically terminated
}

// ExampleCleanupContainer_withOptions demonstrates using CleanupContainer with
// custom termination options.
func ExampleCleanupContainer_withOptions() {
	ctx := context.Background()

	container, err := testcontainers.Run(ctx,
		"postgres:15-alpine",
		testcontainers.WithExposedPorts("5432/tcp"),
	)
	// Cleanup with custom options - remove specific volumes on cleanup
	testcontainers.CleanupContainer(
		&testing.T{},
		container,
		testcontainers.RemoveVolumes("my-volume-name"),
	)
	if err != nil {
		panic(err)
	}

	// Use container...
}

// TestCleanupContainer_realTest shows a complete real test using CleanupContainer.
func TestCleanupContainer_realTest(t *testing.T) {
	ctx := context.Background()

	// Start container
	container, err := testcontainers.Run(ctx,
		"nginx:alpine",
		testcontainers.WithExposedPorts("80/tcp"),
	)
	// Register cleanup immediately - this uses t.Cleanup() internally
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)

	// Get host and port
	host, err := container.Host(ctx)
	require.NoError(t, err)

	mappedPort, err := container.MappedPort(ctx, "80")
	require.NoError(t, err)

	// Use the container in your test
	_ = host
	_ = mappedPort

	// No need to explicitly terminate - CleanupContainer handles it
	// Cleanup runs automatically when test ends
}

// TestCleanupContainer_subTests demonstrates cleanup with subtests.
func TestCleanupContainer_subTests(t *testing.T) {
	ctx := context.Background()

	// Container for parent test
	parentContainer, err := testcontainers.Run(ctx, "nginx:alpine")
	testcontainers.CleanupContainer(t, parentContainer)
	require.NoError(t, err)

	t.Run("SubTest1", func(t *testing.T) {
		// Subtest-specific container
		subContainer, err := testcontainers.Run(ctx, "nginx:alpine")
		testcontainers.CleanupContainer(t, subContainer) // Cleaned up when SubTest1 ends
		require.NoError(t, err)

		// Use subContainer...
	})

	t.Run("SubTest2", func(t *testing.T) {
		// Another subtest-specific container
		subContainer, err := testcontainers.Run(ctx, "nginx:alpine")
		testcontainers.CleanupContainer(t, subContainer) // Cleaned up when SubTest2 ends
		require.NoError(t, err)

		// Use subContainer...
	})

	// parentContainer is cleaned up when main test ends (after all subtests)
}

// TestCleanupContainer_parallel demonstrates cleanup with parallel tests.
func TestCleanupContainer_parallel(t *testing.T) {
	t.Run("Parallel1", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		container, err := testcontainers.Run(ctx, "nginx:alpine")
		testcontainers.CleanupContainer(t, container)
		require.NoError(t, err)

		// Test logic...
	})

	t.Run("Parallel2", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		container, err := testcontainers.Run(ctx, "nginx:alpine")
		testcontainers.CleanupContainer(t, container)
		require.NoError(t, err)

		// Test logic...
	})
	// Both containers are cleaned up independently when their tests complete
}

// ANTI-PATTERN: Don't use defer with TerminateContainer in tests
// This is acceptable for examples (documentation), but tests should use CleanupContainer
func TestCleanupContainer_antiPattern(t *testing.T) {
	ctx := context.Background()

	container, err := testcontainers.Run(ctx, "nginx:alpine")

	// ❌ AVOID: Manual defer in tests
	defer func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	}()

	require.NoError(t, err)

	// ✅ PREFERRED: Use CleanupContainer instead (see other examples)
	// testcontainers.CleanupContainer(t, container)
}

// TestCleanupNetwork demonstrates cleanup for networks.
func TestCleanupNetwork(t *testing.T) {
	ctx := context.Background()

	// Create network using network package
	// Note: GenericNetwork is deprecated but shown for completeness
	//nolint:staticcheck
	network, err := testcontainers.GenericNetwork(ctx, testcontainers.GenericNetworkRequest{
		NetworkRequest: testcontainers.NetworkRequest{
			Name:   "test-network",
			Driver: "bridge",
		},
	})
	testcontainers.CleanupNetwork(t, network) // Similar pattern for networks
	require.NoError(t, err)

	// Use network...
	// Network is automatically removed when test ends
}
