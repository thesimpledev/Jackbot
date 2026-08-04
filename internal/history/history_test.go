package history

import (
	"context"
	"fmt"
	"sync"
	"testing"
)

func TestPrepareMessagesStartsWithSystemPrompt(t *testing.T) {
	s := New(nil, "You are a helpful cat.")
	messages := s.PrepareMessages("test-channel-id")
	if len(messages) != 1 {
		t.Fatalf("expected only the system message, got %d", len(messages))
	}
	if messages[0].Role != "system" || messages[0].Content != "You are a helpful cat." {
		t.Fatalf("unexpected system message: %+v", messages[0])
	}
}

func TestAddMessageEvictsBeyondMaxHistory(t *testing.T) {
	s := New(nil, "prompt")
	ctx := context.Background()
	for i := range maxHistory + 5 {
		s.AddMessage(ctx, "user", "example-user", fmt.Sprintf("message %d", i), "test-channel-id")
	}
	messages := s.PrepareMessages("test-channel-id")
	if len(messages) != maxHistory+1 {
		t.Fatalf("expected %d messages including system, got %d", maxHistory+1, len(messages))
	}
	if messages[1].Content != "message 5" {
		t.Fatalf("expected oldest surviving message to be %q, got %q", "message 5", messages[1].Content)
	}
	if messages[len(messages)-1].Content != fmt.Sprintf("message %d", maxHistory+4) {
		t.Fatalf("unexpected newest message: %q", messages[len(messages)-1].Content)
	}
}

func TestAddMessageConcurrentAccess(t *testing.T) {
	s := New(nil, "prompt")
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			s.AddMessage(ctx, "user", "example-user", fmt.Sprintf("message %d", n), "test-channel-id")
			_ = s.PrepareMessages("test-channel-id")
		}(i)
	}
	wg.Wait()
	messages := s.PrepareMessages("test-channel-id")
	if len(messages) != maxHistory+1 {
		t.Fatalf("expected %d messages including system, got %d", maxHistory+1, len(messages))
	}
}
