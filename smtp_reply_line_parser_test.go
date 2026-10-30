package smtp_reply_line_parser

import (
	"errors"
	"strings"
	"testing"
)

func TestParseSingleLine(t *testing.T) {
	replies, err := Parse(strings.NewReader("250 OK\r\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replies) != 1 {
		t.Fatalf("expected 1 reply, got %d", len(replies))
	}
	if replies[0].Code != 250 {
		t.Errorf("expected code 250, got %d", replies[0].Code)
	}
	if replies[0].Message != "OK" {
		t.Errorf("expected message %q, got %q", "OK", replies[0].Message)
	}
}

func TestParseContinuation(t *testing.T) {
	input := "250-First line\r\n250-Second line\r\n250 Last line\r\n"
	replies, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replies) != 1 {
		t.Fatalf("expected 1 reply, got %d", len(replies))
	}
	if replies[0].Code != 250 {
		t.Errorf("expected code 250, got %d", replies[0].Code)
	}
	want := "First line Second line Last line"
	if replies[0].Message != want {
		t.Errorf("expected message %q, got %q", want, replies[0].Message)
	}
}

func TestParseMultipleReplies(t *testing.T) {
	input := "220 Ready\r\n250 OK\r\n550 Nope\r\n"
	replies, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replies) != 3 {
		t.Fatalf("expected 3 replies, got %d", len(replies))
	}
	expected := []struct {
		code int
		msg  string
	}{
		{220, "Ready"},
		{250, "OK"},
		{550, "Nope"},
	}
	for i, e := range expected {
		if replies[i].Code != e.code {
			t.Errorf("reply %d: expected code %d, got %d", i, e.code, replies[i].Code)
		}
		if replies[i].Message != e.msg {
			t.Errorf("reply %d: expected message %q, got %q", i, e.msg, replies[i].Message)
		}
	}
}

func TestParseUnterminatedReply(t *testing.T) {
	input := "250-First line\r\n250-Second line\r\n"
	replies, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrUnterminatedReply) {
		t.Fatalf("expected ErrUnterminatedReply, got %v", err)
	}
	if len(replies) != 1 {
		t.Fatalf("expected 1 partial reply returned with error, got %d", len(replies))
	}
	if replies[0].Code != 250 {
		t.Errorf("expected code 250, got %d", replies[0].Code)
	}
	want := "First line Second line"
	if replies[0].Message != want {
		t.Errorf("expected message %q, got %q", want, replies[0].Message)
	}
}

func TestParseEmptyInput(t *testing.T) {
	replies, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if replies != nil && len(replies) != 0 {
		t.Fatalf("expected no replies, got %d", len(replies))
	}
}

func TestParseLineTooShort(t *testing.T) {
	_, err := Parse(strings.NewReader("250\r\n"))
	if err == nil {
		t.Fatalf("expected error for line too short")
	}
}

func TestParseNonDigitCode(t *testing.T) {
	_, err := Parse(strings.NewReader("2x0 OK\r\n"))
	if err == nil {
		t.Fatalf("expected error for non-digit code")
	}
}

func TestParseInvalidSeparator(t *testing.T) {
	_, err := Parse(strings.NewReader("250:OK\r\n"))
	if err == nil {
		t.Fatalf("expected error for invalid separator")
	}
}

func TestParseCodeMismatch(t *testing.T) {
	input := "250-First line\r\n251 Second line\r\n"
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatalf("expected error for code mismatch in continuation")
	}
}

func TestParseMessageWithLeadingSpace(t *testing.T) {
	replies, err := Parse(strings.NewReader("250   Extra spaces\r\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replies) != 1 {
		t.Fatalf("expected 1 reply, got %d", len(replies))
	}
	if replies[0].Message != "Extra spaces" {
		t.Errorf("expected trimmed message %q, got %q", "Extra spaces", replies[0].Message)
	}
}

func TestParseUnixLineEndings(t *testing.T) {
	replies, err := Parse(strings.NewReader("250 OK\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replies) != 1 || replies[0].Code != 250 || replies[0].Message != "OK" {
		t.Fatalf("unexpected result: %+v", replies)
	}
}

func TestParseNoTrailingNewline(t *testing.T) {
	replies, err := Parse(strings.NewReader("250 OK"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replies) != 1 || replies[0].Code != 250 || replies[0].Message != "OK" {
		t.Fatalf("unexpected result: %+v", replies)
	}
}

func TestParseMixedContinuationAndComplete(t *testing.T) {
	input := "250-First\r\n250 Second\r\n550 Done\r\n"
	replies, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(replies) != 2 {
		t.Fatalf("expected 2 replies, got %d", len(replies))
	}
	if replies[0].Code != 250 || replies[0].Message != "First Second" {
		t.Errorf("reply 0 mismatch: %+v", replies[0])
	}
	if replies[1].Code != 550 || replies[1].Message != "Done" {
		t.Errorf("reply 1 mismatch: %+v", replies[1])
	}
}
