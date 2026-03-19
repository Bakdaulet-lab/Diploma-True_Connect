package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

// MessageRepo implements repository.MessageRepository using PostgreSQL.
type MessageRepo struct {
	pool *pgxpool.Pool
}

var _ repository.MessageRepository = (*MessageRepo)(nil)

// NewMessageRepo creates a new PostgreSQL-backed message repository.
func NewMessageRepo(pool *pgxpool.Pool) *MessageRepo {
	return &MessageRepo{pool: pool}
}

func (r *MessageRepo) Create(ctx context.Context, msg *domain.Message) error {
	query := `
		INSERT INTO social.messages (match_id, sender_id, content_encrypted)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err := runner(ctx, r.pool).QueryRow(ctx, query,
		msg.MatchID, msg.SenderID, msg.ContentEncrypted,
	).Scan(&msg.ID, &msg.CreatedAt)
	if err != nil {
		return fmt.Errorf("creating message: %w", err)
	}

	return nil
}

func (r *MessageRepo) ListByMatch(ctx context.Context, matchID uuid.UUID, cursor string, limit int) ([]domain.Message, string, error) {
	query := `
                SELECT id, match_id, sender_id, content_encrypted, read_at, created_at
                FROM social.messages
                WHERE match_id = $1 AND ($3::timestamptz IS NULL OR created_at < $3::timestamptz)
                ORDER BY created_at DESC
                LIMIT $2`

	var cursorArg interface{}
	if cursor == "" {
		cursorArg = nil
	} else {
		cursorArg = cursor
	}

	rows, err := runner(ctx, r.pool).Query(ctx, query, matchID, limit, cursorArg)
	if err != nil {
		return nil, "", fmt.Errorf("listing messages: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(
			&m.ID, &m.MatchID, &m.SenderID,
			&m.ContentEncrypted, &m.ReadAt, &m.CreatedAt,
		); err != nil {
			return nil, "", fmt.Errorf("scanning message row: %w", err)
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterating messages: %w", err)
	}

	nextCursor := ""
	if len(messages) == limit {
		nextCursor = messages[len(messages)-1].CreatedAt.Format("2006-01-02T15:04:05.999999Z07:00")
	}

	// Reverse to ASC order for typical chat display (oldest first in the paginated slice)
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nextCursor, nil
}

func (r *MessageRepo) MarkRead(ctx context.Context, matchID uuid.UUID, readerID uuid.UUID) error {
	query := `
		UPDATE social.messages
		SET read_at = NOW()
		WHERE match_id = $1
		  AND sender_id != $2
		  AND read_at IS NULL`

	_, err := runner(ctx, r.pool).Exec(ctx, query, matchID, readerID)
	if err != nil {
		return fmt.Errorf("marking messages read: %w", err)
	}

	return nil
}
