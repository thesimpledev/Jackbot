// Package db persists chat messages and moderation violations to the
// Turso (libsql) database.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// Registers the "libsql" driver with database/sql.
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// Store wraps the Turso (libsql) connection used for the chat and
// violations tables.
type Store struct {
	conn *sql.DB
}

// New opens the libsql connection for the given database URL and token.
func New(url, authToken string) (*Store, error) {
	if url == "" {
		return nil, fmt.Errorf("db: database URL is empty")
	}
	if authToken == "" {
		return nil, fmt.Errorf("db: auth token is empty")
	}
	conn, err := sql.Open("libsql", fmt.Sprintf("%s?authToken=%s", url, authToken))
	if err != nil {
		return nil, fmt.Errorf("db: open: %w", err)
	}
	return &Store{conn: conn}, nil
}

// InsertChatRow appends one message to the chat table.
func (s *Store) InsertChatRow(ctx context.Context, role, speaker, message, roomID string) error {
	if s == nil || s.conn == nil {
		return fmt.Errorf("db: store is not initialized")
	}
	_, err := s.conn.ExecContext(ctx,
		"INSERT INTO chat (Role, Speaker, Message, room, created_at) VALUES (?, ?, ?, ?, ?)",
		role, speaker, message, roomID, formatLocalDateTime(time.Now()))
	if err != nil {
		return fmt.Errorf("db: insert chat row: %w", err)
	}
	return nil
}

// InsertModViolation appends one flagged message to the violations table.
func (s *Store) InsertModViolation(ctx context.Context, speaker, discriminator, message string) error {
	if s == nil || s.conn == nil {
		return fmt.Errorf("db: store is not initialized")
	}
	_, err := s.conn.ExecContext(ctx,
		"INSERT INTO violations (Speaker, Discriminator, Message, created_at) VALUES (?, ?, ?, ?)",
		speaker, discriminator, message, formatLocalDateTime(time.Now()))
	if err != nil {
		return fmt.Errorf("db: insert mod violation: %w", err)
	}
	return nil
}

// formatLocalDateTime matches the created_at format the TypeScript bot
// wrote with Intl.DateTimeFormat('en-US', ...): "01/02/2006, 15:04:05 MST".
func formatLocalDateTime(t time.Time) string {
	return t.Format("01/02/2006, 15:04:05 MST")
}
