package bot

import (
	"context"
	"testing"

	"github.com/bwmarrin/discordgo"
)

type fakeAI struct {
	flagged  bool
	response string
}

func (f *fakeAI) GenerateResponse(_ context.Context, _ string) string { return f.response }
func (f *fakeAI) IsFlagged(_ context.Context, _ string) bool          { return f.flagged }

type addedMessage struct {
	role, speaker, content, channelID string
}

type fakeHistory struct {
	added []addedMessage
}

func (f *fakeHistory) AddMessage(_ context.Context, role, speaker, content, channelID string) {
	f.added = append(f.added, addedMessage{role, speaker, content, channelID})
}

type violationRow struct {
	speaker, discriminator, message string
}

type fakeViolations struct {
	rows []violationRow
}

func (f *fakeViolations) InsertModViolation(_ context.Context, speaker, discriminator, message string) error {
	f.rows = append(f.rows, violationRow{speaker, discriminator, message})
	return nil
}

type fakeResponder struct {
	replies []string
}

func (f *fakeResponder) ChannelMessageSendReply(_ string, content string, _ *discordgo.MessageReference, _ ...discordgo.RequestOption) (*discordgo.Message, error) {
	f.replies = append(f.replies, content)
	return &discordgo.Message{}, nil
}

func directMessage() *discordgo.Message {
	return &discordgo.Message{
		Author:    &discordgo.User{Bot: false, GlobalName: "example-user", Discriminator: "0001"},
		Content:   "Hello, Bot!",
		GuildID:   "",
		ChannelID: "test-channel-id",
	}
}

func TestHandleRespondsToDirectMessage(t *testing.T) {
	ai := &fakeAI{response: "Mock AI response"}
	hist := &fakeHistory{}
	violations := &fakeViolations{}
	responder := &fakeResponder{}
	h := NewHandler(ai, hist, violations, "Bot", "cat")

	h.handle(context.Background(), responder, directMessage())

	if len(responder.replies) != 1 || responder.replies[0] != "Mock AI response" {
		t.Fatalf("expected one reply %q, got %v", "Mock AI response", responder.replies)
	}
	if len(hist.added) != 1 || hist.added[0].role != "user" || hist.added[0].speaker != "example-user" {
		t.Fatalf("expected the user message in history, got %v", hist.added)
	}
	if len(violations.rows) != 0 {
		t.Fatalf("expected no violations, got %v", violations.rows)
	}
}

func TestHandleIgnoresBotAuthors(t *testing.T) {
	responder := &fakeResponder{}
	h := NewHandler(&fakeAI{response: "Mock AI response"}, &fakeHistory{}, &fakeViolations{}, "Bot", "cat")

	m := directMessage()
	m.Author.Bot = true
	h.handle(context.Background(), responder, m)

	if len(responder.replies) != 0 {
		t.Fatalf("expected no replies, got %v", responder.replies)
	}
}

func TestHandleFlaggedMessage(t *testing.T) {
	hist := &fakeHistory{}
	violations := &fakeViolations{}
	responder := &fakeResponder{}
	h := NewHandler(&fakeAI{flagged: true}, hist, violations, "Bot", "cat")

	h.handle(context.Background(), responder, directMessage())

	want := "Nope, cat can't process that type of content"
	if len(responder.replies) != 1 || responder.replies[0] != want {
		t.Fatalf("expected refusal %q, got %v", want, responder.replies)
	}
	if len(violations.rows) != 1 || violations.rows[0].speaker != "example-user" {
		t.Fatalf("expected one violation for example-user, got %v", violations.rows)
	}
	if len(hist.added) != 0 {
		t.Fatalf("expected nothing in history, got %v", hist.added)
	}
}
