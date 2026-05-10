package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

// MahramChatRepo implements repository.MahramChatRepository using PostgreSQL.
type MahramChatRepo struct {
	pool *pgxpool.Pool
}

var _ repository.MahramChatRepository = (*MahramChatRepo)(nil)

// NewMahramChatRepo creates a new PostgreSQL-backed mahram chat repository.
func NewMahramChatRepo(pool *pgxpool.Pool) *MahramChatRepo {
	return &MahramChatRepo{pool: pool}
}

func (r *MahramChatRepo) CreateRoom(ctx context.Context, matchID, mahramUserID uuid.UUID) (*domain.MahramChatRoom, error) {
	const q = `
		INSERT INTO social.mahram_chat_rooms (match_id, mahram_user_id)
		VALUES ($1, $2)
		ON CONFLICT (match_id) DO UPDATE SET mahram_user_id = EXCLUDED.mahram_user_id
		RETURNING id, created_at`

	rm := &domain.MahramChatRoom{MatchID: matchID, MahramUserID: mahramUserID}
	if err := runner(ctx, r.pool).QueryRow(ctx, q, matchID, mahramUserID).Scan(&rm.ID, &rm.CreatedAt); err != nil {
		return nil, fmt.Errorf("creating mahram chat room: %w", err)
	}
	if err := r.loadParticipants(ctx, rm); err != nil {
		return nil, err
	}
	return rm, nil
}

func (r *MahramChatRepo) GetRoomByMatchID(ctx context.Context, matchID uuid.UUID) (*domain.MahramChatRoom, error) {
	const q = `
		SELECT r.id, r.match_id, m.user_a_id, m.user_b_id, r.mahram_user_id, r.created_at
		FROM social.mahram_chat_rooms r
		JOIN social.matches m ON m.id = r.match_id
		WHERE r.match_id = $1`

	rm := &domain.MahramChatRoom{}
	err := runner(ctx, r.pool).QueryRow(ctx, q, matchID).Scan(
		&rm.ID, &rm.MatchID, &rm.UserAID, &rm.UserBID, &rm.MahramUserID, &rm.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("mahram chat room by match: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("mahram chat room by match: %w", err)
	}
	return rm, nil
}

func (r *MahramChatRepo) GetRoomByID(ctx context.Context, roomID uuid.UUID) (*domain.MahramChatRoom, error) {
	const q = `
		SELECT r.id, r.match_id, m.user_a_id, m.user_b_id, r.mahram_user_id, r.created_at
		FROM social.mahram_chat_rooms r
		JOIN social.matches m ON m.id = r.match_id
		WHERE r.id = $1`

	rm := &domain.MahramChatRoom{}
	err := runner(ctx, r.pool).QueryRow(ctx, q, roomID).Scan(
		&rm.ID, &rm.MatchID, &rm.UserAID, &rm.UserBID, &rm.MahramUserID, &rm.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("mahram chat room by id: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("mahram chat room by id: %w", err)
	}
	return rm, nil
}

func (r *MahramChatRepo) SaveMessage(ctx context.Context, roomID, senderID uuid.UUID, contentEncrypted []byte) (*domain.MahramMessage, error) {
	const q = `
		INSERT INTO social.mahram_messages (room_id, sender_id, content_encrypted)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	msg := &domain.MahramMessage{RoomID: roomID, SenderID: senderID, ContentEncrypted: contentEncrypted}
	if err := runner(ctx, r.pool).QueryRow(ctx, q, roomID, senderID, contentEncrypted).Scan(&msg.ID, &msg.CreatedAt); err != nil {
		return nil, fmt.Errorf("saving mahram message: %w", err)
	}
	return msg, nil
}

func (r *MahramChatRepo) GetMessages(ctx context.Context, roomID uuid.UUID, cursor string, limit int) ([]*domain.MahramMessage, string, error) {
	var (
		q    string
		args []interface{}
	)
	if cursor == "" {
		q = `
			SELECT id, room_id, sender_id, content_encrypted, created_at
			FROM social.mahram_messages
			WHERE room_id = $1
			ORDER BY created_at DESC
			LIMIT $2`
		args = []interface{}{roomID, limit}
	} else {
		q = `
			SELECT id, room_id, sender_id, content_encrypted, created_at
			FROM social.mahram_messages
			WHERE room_id = $1
			  AND created_at < (SELECT created_at FROM social.mahram_messages WHERE id = $3::uuid)
			ORDER BY created_at DESC
			LIMIT $2`
		args = []interface{}{roomID, limit, cursor}
	}

	rows, err := runner(ctx, r.pool).Query(ctx, q, args...)
	if err != nil {
		return nil, "", fmt.Errorf("getting mahram messages: %w", err)
	}
	defer rows.Close()

	var msgs []*domain.MahramMessage
	for rows.Next() {
		msg := &domain.MahramMessage{}
		if err := rows.Scan(&msg.ID, &msg.RoomID, &msg.SenderID, &msg.ContentEncrypted, &msg.CreatedAt); err != nil {
			return nil, "", fmt.Errorf("scanning mahram message: %w", err)
		}
		msgs = append(msgs, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterating mahram messages: %w", err)
	}

	nextCursor := ""
	if len(msgs) == limit {
		nextCursor = msgs[len(msgs)-1].ID.String()
	}
	return msgs, nextCursor, nil
}

func (r *MahramChatRepo) loadParticipants(ctx context.Context, rm *domain.MahramChatRoom) error {
	const q = `SELECT user_a_id, user_b_id FROM social.matches WHERE id = $1`
	if err := runner(ctx, r.pool).QueryRow(ctx, q, rm.MatchID).Scan(&rm.UserAID, &rm.UserBID); err != nil {
		return fmt.Errorf("loading match participants for mahram room: %w", err)
	}
	return nil
}
