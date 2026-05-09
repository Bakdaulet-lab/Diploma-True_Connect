package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

type NotificationRepo struct {
	pool *pgxpool.Pool
}

var _ repository.NotificationRepository = (*NotificationRepo)(nil)

func NewNotificationRepo(pool *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{pool: pool}
}

func (r *NotificationRepo) Create(ctx context.Context, notif *domain.Notification) error {
	query := `
		INSERT INTO social.notifications (user_id, actor_id, type, entity_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, is_read, created_at
	`
	err := runner(ctx, r.pool).QueryRow(ctx, query, notif.UserID, notif.ActorID, notif.Type, notif.EntityID).
		Scan(&notif.ID, &notif.IsRead, &notif.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert notification: %w", err)
	}
	return nil
}

func (r *NotificationRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	query := `
		SELECT n.id, n.user_id, n.actor_id, n.type, n.entity_id, n.is_read, n.created_at,
		       COALESCE(p.display_name, ''), COALESCE(p.avatar_url, '')
		FROM social.notifications n
		LEFT JOIN social.profiles p ON n.actor_id = p.user_id
		WHERE n.id = $1
	`
	var notif domain.Notification
	err := runner(ctx, r.pool).QueryRow(ctx, query, id).Scan(
		&notif.ID, &notif.UserID, &notif.ActorID, &notif.Type, &notif.EntityID, &notif.IsRead, &notif.CreatedAt,
		&notif.ActorName, &notif.ActorAvatar,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get notification: %w", err)
	}
	return &notif, nil
}

func (r *NotificationRepo) ListByUser(ctx context.Context, userID uuid.UUID, cursor string, limit int) ([]domain.Notification, string, error) {
	var query string
	var rows pgx.Rows
	var err error

	if cursor == "" {
		query = `
			SELECT n.id, n.user_id, n.actor_id, n.type, n.entity_id, n.is_read, n.created_at,
			       COALESCE(p.display_name, ''), COALESCE(p.avatar_url, '')
			FROM social.notifications n
			LEFT JOIN social.profiles p ON n.actor_id = p.user_id
			WHERE n.user_id = $1
			ORDER BY n.created_at DESC
			LIMIT $2
		`
		rows, err = runner(ctx, r.pool).Query(ctx, query, userID, limit)
	} else {
		query = `
			SELECT n.id, n.user_id, n.actor_id, n.type, n.entity_id, n.is_read, n.created_at,
			       COALESCE(p.display_name, ''), COALESCE(p.avatar_url, '')
			FROM social.notifications n
			LEFT JOIN social.profiles p ON n.actor_id = p.user_id
			WHERE n.user_id = $1 AND n.created_at < $3::timestamptz
			ORDER BY n.created_at DESC
			LIMIT $2
		`
		rows, err = runner(ctx, r.pool).Query(ctx, query, userID, limit, cursor)
	}

	if err != nil {
		return nil, "", fmt.Errorf("listing notifications: %w", err)
	}
	defer rows.Close()

	var notifs []domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(
			&n.ID, &n.UserID, &n.ActorID, &n.Type, &n.EntityID, &n.IsRead, &n.CreatedAt,
			&n.ActorName, &n.ActorAvatar,
		); err != nil {
			return nil, "", fmt.Errorf("scanning notification: %w", err)
		}
		notifs = append(notifs, n)
	}

	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterating notifications: %w", err)
	}

	var nextCursor string
	if len(notifs) == limit {
		nextCursor = notifs[len(notifs)-1].CreatedAt.Format("2006-01-02T15:04:05.999999Z07:00")
	}

	return notifs, nextCursor, nil
}

func (r *NotificationRepo) MarkAsRead(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	cmdTag, err := runner(ctx, r.pool).Exec(ctx, `UPDATE social.notifications SET is_read = TRUE WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("mark read: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *NotificationRepo) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	_, err := runner(ctx, r.pool).Exec(ctx, `UPDATE social.notifications SET is_read = TRUE WHERE user_id = $1 AND is_read = FALSE`, userID)
	if err != nil {
		return fmt.Errorf("mark all read: %w", err)
	}
	return nil
}

func (r *NotificationRepo) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := runner(ctx, r.pool).QueryRow(ctx, `SELECT COUNT(*) FROM social.notifications WHERE user_id = $1 AND is_read = FALSE`, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count unread: %w", err)
	}
	return count, nil
}
