package bot

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSplitTelegramTextKeepsMessagesUnderLimit(t *testing.T) {
	text := strings.Repeat("metric=value\n", 500)
	chunks := splitTelegramText(text, 4000)
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}
	var combined strings.Builder
	for _, chunk := range chunks {
		if utf8.RuneCountInString(chunk) > 4000 {
			t.Fatalf("chunk exceeds Telegram limit: %d", utf8.RuneCountInString(chunk))
		}
		combined.WriteString(chunk)
	}
	if combined.String() != text {
		t.Fatal("split chunks do not reconstruct original text")
	}
}

func TestSplitTelegramTextDoesNotSplitUnicodeRune(t *testing.T) {
	text := strings.Repeat("監控服務", 2000)
	chunks := splitTelegramText(text, 4000)
	for _, chunk := range chunks {
		if !utf8.ValidString(chunk) {
			t.Fatal("chunk contains invalid UTF-8")
		}
	}
}
