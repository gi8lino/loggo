package ingest

import (
	"context"
	"strings"
	"testing"
	"time"
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

	if len(events) != 2 {
		t.Fatalf("expected 2 logical events, got %d", len(events))
	}

	if !strings.Contains(events[0], "NullPointerException") || !strings.Contains(events[0], "Caused by:") {
		t.Fatalf("expected first event to include stacktrace lines, got %q", events[0])
	}

	if events[1] != `2026-05-13 10:15:31 INFO OrderService recovered` {
		t.Fatalf("unexpected second event: %q", events[1])
	}
}
