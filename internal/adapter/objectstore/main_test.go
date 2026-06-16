package objectstore

import (
	"context"
	"fmt"
	"os"
	"testing"

	tcminio "github.com/testcontainers/testcontainers-go/modules/minio"
)

var testStore *PictureStore

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()

	container, err := tcminio.Run(ctx, "minio/minio:RELEASE.2024-01-16T16-07-38Z")
	if err != nil {
		// In CI a failure to start the container is a real failure (a false-green run would
		// hide the untested object store). Locally, where Docker may be absent, skip instead.
		if os.Getenv("CI") != "" {
			fmt.Fprintln(os.Stderr, "minio integration tests failed to start in CI:", err)
			return 1
		}
		fmt.Fprintln(os.Stderr, "skipping minio integration tests (Docker unavailable):", err)
		return m.Run()
	}
	defer func() { _ = container.Terminate(ctx) }()

	endpoint, err := container.ConnectionString(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connection string:", err)
		return 1
	}

	store, err := New(endpoint, container.Username, container.Password, "pet-pictures-test", false)
	if err != nil {
		fmt.Fprintln(os.Stderr, "new store:", err)
		return 1
	}
	if err := store.EnsureBucket(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "ensure bucket:", err)
		return 1
	}
	testStore = store

	return m.Run()
}

func requireStore(t *testing.T) {
	t.Helper()
	if testStore == nil {
		t.Skip("minio container unavailable")
	}
}
