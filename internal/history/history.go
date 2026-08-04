// Package history keeps the recent conversation per channel in memory
// and mirrors every message into the chat store.
package history

import (
	"context"
	"log"
	"sync"
)

// Message is one entry of a channel's conversation history.
type Message struct {
	Role    string
	Speaker string
	Content string
}

// ChatStore persists chat rows; satisfied by *db.Store.
type ChatStore interface {
	InsertChatRow(ctx context.Context, role, speaker, message, roomID string) error
}

const maxHistory = 25

// Service keeps the last maxHistory messages per channel in memory and
// mirrors every message into the chat store. Discord handlers run
// concurrently, so access is mutex-guarded.
type Service struct {
	mu           sync.Mutex
	store        ChatStore
	systemPrompt string
	channels     map[string][]Message
}

// New builds a Service backed by store; an empty systemPrompt falls
// back to the same default the TypeScript bot used.
func New(store ChatStore, systemPrompt string) *Service {
	if systemPrompt == "" {
		systemPrompt = "Default prompt"
	}
	return &Service{
		store:        store,
		systemPrompt: systemPrompt,
		channels:     make(map[string][]Message),
	}
}

// AddMessage appends a message to the channel's history, evicting the
// oldest entry at capacity, and mirrors it into the chat store.
func (s *Service) AddMessage(ctx context.Context, role, speaker, content, channelID string) {
	if s == nil || channelID == "" {
		return
	}
	s.mu.Lock()
	messages := s.channels[channelID]
	if len(messages) >= maxHistory {
		messages = messages[1:]
	}
	s.channels[channelID] = append(messages, Message{Role: role, Speaker: speaker, Content: content})
	s.mu.Unlock()

	if s.store == nil {
		return
	}
	if err := s.store.InsertChatRow(ctx, role, speaker, content, channelID); err != nil {
		log.Printf("failed to insert chat row: %v", err)
	}
}

// PrepareMessages returns the system prompt followed by the channel's
// history, ready to send to the model.
func (s *Service) PrepareMessages(channelID string) []Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	channelMessages := s.channels[channelID]
	messages := make([]Message, 0, len(channelMessages)+1)
	messages = append(messages, Message{Role: "system", Speaker: "System", Content: s.systemPrompt})
	messages = append(messages, channelMessages...)
	return messages
}
