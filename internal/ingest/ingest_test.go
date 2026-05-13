package ingest

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartJoinsJavaStacktracesIntoOneEvent(t *testing.T) {
	input := strings.Join([]string{
		`2026-05-13 10:15:30 ERROR OrderService Failed to process order`,
		`java.lang.NullPointerException: customer was nil`,
		`    at com.example.OrderService.run(OrderService.java:42)`,
		`Caused by: java.lang.IllegalStateException: missing state`,
		`    at com.example.State.load(State.java:17)`,
		`2026-05-13 10:15:31 INFO OrderService recovered`,
	}, "\n")

	stream := Start(context.Background(), strings.NewReader(input), Options{
		BatchSize:     8,
		FlushInterval: time.Millisecond,
		JoinMultiline: true,
		MaxEventLines: 32,
	})

	events := []string{}
	for batch := range stream {
		events = append(events, batch.Lines...)
	}

	require.Len(t, events, 2)
	assert.Contains(t, events[0], "NullPointerException")
	assert.Contains(t, events[0], "Caused by:")
	assert.Equal(t, `2026-05-13 10:15:31 INFO OrderService recovered`, events[1])
}
