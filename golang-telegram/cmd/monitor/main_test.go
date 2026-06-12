package main

import (
	"os"
	"syscall"
	"testing"
)

func TestSignalReason(t *testing.T) {
	if got := signalReason(os.Interrupt); got != "收到中斷訊號（SIGINT / Ctrl+C）" {
		t.Fatalf("unexpected interrupt reason: %q", got)
	}
	if got := signalReason(syscall.SIGTERM); got != "收到終止訊號（SIGTERM）" {
		t.Fatalf("unexpected SIGTERM reason: %q", got)
	}
}
