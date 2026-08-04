// Package bot connects to the Discord gateway and routes incoming
// messages through moderation, history, and response generation.
package bot

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/bwmarrin/discordgo"
)

// Responder is the slice of *discordgo.Session the handler needs to
// reply, kept narrow so tests can fake it.
type Responder interface {
	ChannelMessageSendReply(channelID string, content string, reference *discordgo.MessageReference, options ...discordgo.RequestOption) (*discordgo.Message, error)
}

// AI generates replies and moderation verdicts.
type AI interface {
	GenerateResponse(ctx context.Context, channelID string) string
	IsFlagged(ctx context.Context, content string) bool
}

// History records conversation messages.
type History interface {
	AddMessage(ctx context.Context, role, speaker, content, channelID string)
}

// ViolationStore persists moderation violations.
type ViolationStore interface {
	InsertModViolation(ctx context.Context, speaker, discriminator, message string) error
}

// Handler processes incoming Discord messages.
type Handler struct {
	ai         AI
	history    History
	violations ViolationStore
	botName    *regexp.Regexp
	animal     string
}

// NewHandler wires the handler's dependencies. botName becomes the
// case-insensitive trigger pattern for guild messages.
func NewHandler(ai AI, history History, violations ViolationStore, botName, animal string) *Handler {
	return &Handler{
		ai:         ai,
		history:    history,
		violations: violations,
		botName:    regexp.MustCompile("(?i)" + regexp.QuoteMeta(botName)),
		animal:     animal,
	}
}

// HandleMessage is the discordgo messageCreate handler.
func (h *Handler) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	h.handle(context.Background(), s, m.Message)
}

func (h *Handler) handle(ctx context.Context, r Responder, m *discordgo.Message) {
	if h == nil || m == nil || m.Author == nil {
		return
	}
	if m.Author.Bot || (!h.botName.MatchString(m.Content) && m.GuildID != "") {
		return
	}
	if m.Content == "" || m.ChannelID == "" || m.Author.GlobalName == "" || m.Author.Discriminator == "" {
		log.Printf("malformed message object")
		return
	}
	if h.ai.IsFlagged(ctx, m.Content) {
		h.handleFlagged(ctx, r, m)
		return
	}
	h.history.AddMessage(ctx, "user", m.Author.GlobalName, m.Content, m.ChannelID)
	reply := h.ai.GenerateResponse(ctx, m.ChannelID)
	log.Println(reply)
	if _, err := r.ChannelMessageSendReply(m.ChannelID, reply, m.Reference()); err != nil {
		log.Printf("failed to send reply: %v", err)
	}
}

func (h *Handler) handleFlagged(ctx context.Context, r Responder, m *discordgo.Message) {
	refusal := fmt.Sprintf("Nope, %s can't process that type of content", h.animal)
	if _, err := r.ChannelMessageSendReply(m.ChannelID, refusal, m.Reference()); err != nil {
		log.Printf("failed to send refusal: %v", err)
	}
	log.Printf("Moderator flagged message: %s %s %s", m.Author.GlobalName, m.Author.Discriminator, m.Content)
	if err := h.violations.InsertModViolation(ctx, m.Author.GlobalName, m.Author.Discriminator, m.Content); err != nil {
		log.Printf("failed to insert mod flag: %v", err)
	}
}

// New opens a Discord session with the same intents as the TypeScript
// bot and wires the handler. The caller closes the returned session.
func New(token, botName string, h *Handler) (*discordgo.Session, error) {
	if token == "" {
		return nil, fmt.Errorf("bot: discord token is empty")
	}
	if h == nil {
		return nil, fmt.Errorf("bot: handler is nil")
	}
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("bot: create session: %w", err)
	}
	session.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentMessageContent |
		discordgo.IntentsDirectMessages |
		discordgo.IntentsDirectMessageReactions
	session.AddHandlerOnce(func(_ *discordgo.Session, _ *discordgo.Ready) {
		log.Printf("%s is ready!", botName)
	})
	session.AddHandler(h.HandleMessage)
	if err := session.Open(); err != nil {
		return nil, fmt.Errorf("bot: open session: %w", err)
	}
	return session, nil
}
