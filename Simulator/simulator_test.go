package Simulator

import (
	"testing"
	"time"

	"github.com/HadasAmar/analytics-load-tool/Model"
)

func TestCalculateReplayEvents(t *testing.T) {
	records := []*Model.ParsedRecord{
		{
			LogTime: parseTime(t, "2024-05-01T08:00:00Z"),
			IP:      "1.1.1.1",
		},
		{
			LogTime: parseTime(t, "2024-05-01T08:00:07Z"),
			IP:      "1.1.1.2",
		},
		{
			LogTime: parseTime(t, "2024-05-01T08:00:17Z"),
			IP:      "1.1.1.3",
		},
	}

	events, err := CalculateReplayEvents(records)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	if events[0].Delay != 0 {
		t.Errorf("first event delay should be 0, got %v", events[0].Delay)
	}

	if events[1].Delay != 7*time.Second {
		t.Errorf("second event delay should be 7s, got %v", events[1].Delay)
	}

	if events[2].Delay != 10*time.Second {
		t.Errorf("third event delay should be 10s, got %v", events[2].Delay)
	}
}

func parseTime(t *testing.T, s string) time.Time {
	tt, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("invalid time format: %v", err)
	}
	return tt
}

func TestReplaySpeedup(t *testing.T) {
	cases := []struct {
		delay   time.Duration
		speedup float64
		expect  time.Duration
	}{
		{10 * time.Second, 2.0, 5 * time.Second},
		{5 * time.Second, 0.0, 5 * time.Second}, // speedup=0 אמור להפוך ל-1
		{5 * time.Second, 1.0, 5 * time.Second},
		{1000 * time.Millisecond, 4.0, 250 * time.Millisecond},
	}

	for _, c := range cases {
		got := ReplaySpeedup(c.delay, c.speedup)
		if got != c.expect {
			t.Errorf("ReplaySpeedup(%v, %v) = %v; want %v", c.delay, c.speedup, got, c.expect)
		}
	}
}

func TestSimulateReplayWithoutPause(t *testing.T) {
    records := []*Model.ParsedRecord{
        {LogTime: parseTime(t, "2024-05-01T10:00:00Z"), IP: "1.1.1.1"},
        {LogTime: parseTime(t, "2024-05-01T10:00:01Z"), IP: "1.1.1.2"},
        {LogTime: parseTime(t, "2024-05-01T10:00:02Z"), IP: "1.1.1.3"},
    }

    commands := make(chan string)

    start := time.Now()
    err := SimulateReplayWithControl(records, commands)
    if err != nil {
        t.Errorf("error: %v", err)
    }
    elapsed := time.Since(start)
    if elapsed < time.Second {
        t.Errorf("expected a delay, got %v", elapsed)
    }
}

func TestSimulateReplayWithPauseResume(t *testing.T) {
    records := []*Model.ParsedRecord{
        {LogTime: parseTime(t, "2024-05-01T10:00:00Z"), IP: "1.1.1.1"},
        {LogTime: parseTime(t, "2024-05-01T10:00:01Z"), IP: "1.1.1.2"},
        {LogTime: parseTime(t, "2024-05-01T10:00:02Z"), IP: "1.1.1.3"},
    }

    commands := make(chan string)

    go func() {
        time.Sleep(100 * time.Millisecond)
        commands <- "pause"
        t.Log("sent pause")
        time.Sleep(300 * time.Millisecond)
        commands <- "resume"
        t.Log("sent resume")
        time.Sleep(200 * time.Millisecond)
        commands <- "stop"
        t.Log("sent stop")
    }()

    err := SimulateReplayWithControl(records, commands)
    if err != nil {
        t.Errorf("error: %v", err)
    }
}
