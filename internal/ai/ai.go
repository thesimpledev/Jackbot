// Package ai generates chat replies and moderation verdicts through
// the OpenAI API.
package ai

import (
	"context"
	"fmt"
	"log"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"

	"github.com/thesimpledev/Jackbot/internal/history"
)

const fallbackResponse = "Sorry, I'm having a pawsitively hard time understanding you right meow. Please try again later."

// Service generates chat responses and moderation verdicts via OpenAI,
// reading and writing conversation state through the history service.
type Service struct {
	client  openai.Client
	model   string
	botName string
	hist    *history.Service
}

// New builds a Service; empty model and botName fall back to the same
// defaults the TypeScript bot used.
func New(apiKey, model, botName string, hist *history.Service) *Service {
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	if botName == "" {
		botName = "Bot"
	}
	return &Service{
		client:  openai.NewClient(option.WithAPIKey(apiKey)),
		model:   model,
		botName: botName,
		hist:    hist,
	}
}

// GenerateResponse asks the model for a reply to the channel's history.
// Errors never escape: on any failure the fallback text is recorded and
// returned, matching the TypeScript bot.
func (s *Service) GenerateResponse(ctx context.Context, channelID string) string {
	if s == nil || s.hist == nil {
		return fallbackResponse
	}
	responseMessage, err := s.complete(ctx, s.hist.PrepareMessages(channelID))
	if err != nil {
		log.Printf("Error: %v", err)
		s.hist.AddMessage(ctx, "assistant", s.botName, fallbackResponse, channelID)
		return fallbackResponse
	}
	s.hist.AddMessage(ctx, "assistant", s.botName, responseMessage, channelID)
	return responseMessage
}

func (s *Service) complete(ctx context.Context, messages []history.Message) (string, error) {
	params := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))
	for _, m := range messages {
		params = append(params, toParam(m))
	}
	completion, err := s.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(s.model),
		Messages: params,
	})
	if err != nil {
		return "", err
	}
	if len(completion.Choices) == 0 || completion.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("no response message")
	}
	return completion.Choices[0].Message.Content, nil
}

// IsFlagged reports whether the moderation endpoint flags the content.
// API errors are logged and treated as not flagged, matching the
// TypeScript bot.
func (s *Service) IsFlagged(ctx context.Context, content string) bool {
	if s == nil {
		return false
	}
	moderation, err := s.client.Moderations.New(ctx, openai.ModerationNewParams{
		Input: openai.ModerationNewParamsInputUnion{OfString: openai.String(content)},
	})
	if err != nil {
		log.Printf("Moderation error: %v", err)
		return false
	}
	if len(moderation.Results) == 0 {
		return false
	}
	return moderation.Results[0].Flagged
}

func toParam(m history.Message) openai.ChatCompletionMessageParamUnion {
	switch m.Role {
	case "system":
		return openai.ChatCompletionMessageParamUnion{
			OfSystem: &openai.ChatCompletionSystemMessageParam{
				Content: openai.ChatCompletionSystemMessageParamContentUnion{OfString: openai.String(m.Content)},
				Name:    openai.String(m.Speaker),
			},
		}
	case "assistant":
		return openai.ChatCompletionMessageParamUnion{
			OfAssistant: &openai.ChatCompletionAssistantMessageParam{
				Content: openai.ChatCompletionAssistantMessageParamContentUnion{OfString: openai.String(m.Content)},
				Name:    openai.String(m.Speaker),
			},
		}
	default:
		return openai.ChatCompletionMessageParamUnion{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{OfString: openai.String(m.Content)},
				Name:    openai.String(m.Speaker),
			},
		}
	}
}
