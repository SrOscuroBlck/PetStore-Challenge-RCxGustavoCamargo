package rediscache

import (
	"context"
	"fmt"
	"os"
	"testing"

	"roboticCrewChallenge/internal/testsupport"
)

var harness *testsupport.Harness

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	h, cleanup, ok, err := testsupport.Start(context.Background(), testsupport.Options{WithRedis: true})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if !ok {
		return m.Run()
	}
	defer cleanup()
	harness = h
	return m.Run()
}

func requireRedis(t *testing.T) {
	t.Helper()
	if harness == nil || harness.RedisClient == nil {
		t.Skip("redis container unavailable")
	}
}
