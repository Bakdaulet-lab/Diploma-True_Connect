package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trueconnect/backend/internal/domain"
	"github.com/trueconnect/backend/internal/repository"
)

// PostRepo implements repository.PostRepository using PostgreSQL.
type PostRepo struct {
	pool *pgxpool.Pool
}

var _ repository.PostRepository = (*PostRepo)(nil)

// NewPostRepo creates a new PostgreSQL-backed post repository.
func NewPostRepo(pool *pgxpool.Pool) *PostRepo {
	return &PostRepo{pool: pool}
}

func (r *PostRepo) Create(ctx context.Context, post *domain.Post) error {
	query := `
		INSERT INTO social.posts (author_id, content, media_url)
		VALUES ($1, $2, $3)
		RETURNING id, like_count, comment_count, created_at, updated_at`

	err := runner(ctx, r.pool).QueryRow(ctx, query,
		post.AuthorID, post.Content, post.MediaURL,
	).Scan(&post.ID, &post.LikeCount, &post.CommentCount, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating post: %w", err)
	}

	return nil
}

func (r *PostRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	query := `
		SELECT p.id, p.author_id, COALESCE(p.content, ''), COALESCE(p.media_url, ''), p.like_count, p.comment_count, p.created_at, p.updated_at,
		       COALESCE(pr.display_name, 'Unknown User'), COALESCE(pr.avatar_url, '')
		FROM social.posts p
		LEFT JOIN social.profiles pr ON p.author_id = pr.user_id
		WHERE p.id = $1`

	p := &domain.Post{}
	err := runner(ctx, r.pool).QueryRow(ctx, query, id).Scan(
		&p.ID, &p.AuthorID, &p.Content, &p.MediaURL,
		&p.LikeCount, &p.CommentCount, &p.CreatedAt, &p.UpdatedAt,
		&p.AuthorName, &p.AuthorAvatar,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("getting post: %w", domain.ErrNotFound)
		}
		return nil, fmt.Errorf("getting post: %w", err)
	}

	return p, nil
}

func (r *PostRepo) Delete(ctx context.Context, id uuid.UUID, authorID uuid.UUID) error {
	query := `DELETE FROM social.posts WHERE id = $1 AND author_id = $2`

	tag, err := runner(ctx, r.pool).Exec(ctx, query, id, authorID)
	if err != nil {
		return fmt.Errorf("deleting post: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("deleting post: %w", domain.ErrNotFound)
	}

	return nil
}

func (r *PostRepo) ListFeed(ctx context.Context, cursor string, limit int, filter domain.PostFilter) ([]domain.Post, string, error) {
	var args []interface{}
	
	baseQuery := `
		SELECT p.id, p.author_id, COALESCE(p.content, ''), COALESCE(p.media_url, ''), p.like_count, p.comment_count, p.created_at, p.updated_at,
		       COALESCE(pr.display_name, 'Unknown User'), COALESCE(pr.avatar_url, '')
		FROM social.posts p
		LEFT JOIN social.profiles pr ON p.author_id = pr.user_id
		WHERE 1=1
	`
	argIdx := 1

	if filter.SearchQuery != "" {
		baseQuery += fmt.Sprintf(" AND p.search_vector @@ plainto_tsquery('english', $%d) ", argIdx)
		args = append(args, filter.SearchQuery)
		argIdx++
	}

	if filter.Timeframe != "" && filter.Timeframe != "all" {
		switch filter.Timeframe {
		case "24h":
			baseQuery += " AND p.created_at > (NOW() - interval '24 hours') "
		case "7d":
			baseQuery += " AND p.created_at > (NOW() - interval '7 days') "
		case "30d":
			baseQuery += " AND p.created_at > (NOW() - interval '30 days') "
		}
	}

	if cursor != "" && filter.SortBy != "popular" {
		baseQuery += fmt.Sprintf(" AND p.created_at < $%d::timestamptz ", argIdx)
		args = append(args, cursor)
		argIdx++
	}

	if filter.SortBy == "popular" {
		baseQuery += " ORDER BY p.like_count DESC, p.created_at DESC "
		if cursor != "" {
			offset, _ := strconv.Atoi(cursor)
			baseQuery += fmt.Sprintf(" OFFSET $%d ", argIdx)
			args = append(args, offset)
			argIdx++
		}
	} else {
		baseQuery += " ORDER BY p.created_at DESC "
	}

	baseQuery += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := runner(ctx, r.pool).Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, "", fmt.Errorf("listing feed: %w", err)
	}
	defer rows.Close()

	var posts []domain.Post
	for rows.Next() {
		var p domain.Post
		if err := rows.Scan(
			&p.ID, &p.AuthorID, &p.Content, &p.MediaURL,
			&p.LikeCount, &p.CommentCount, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorName, &p.AuthorAvatar,
		); err != nil {
			return nil, "", fmt.Errorf("scanning post row: %w", err)
		}
		posts = append(posts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("iterating feed: %w", err)
	}

	var nextCursor string
	if len(posts) == limit {
		if filter.SortBy == "popular" {
			offset := 0
			if cursor != "" {
				offset, _ = strconv.Atoi(cursor)
			}
			offset += limit
			nextCursor = strconv.Itoa(offset)
		} else {
			nextCursor = posts[len(posts)-1].CreatedAt.Format("2006-01-02T15:04:05.999999Z07:00")
		}
	}

	return posts, nextCursor, nil
}

func (r *PostRepo) ListByAuthor(ctx context.Context, authorID uuid.UUID, cursor string, limit int) ([]domain.Post, string, error) {
	var query string
	var rows pgx.Rows
	var err error

	if cursor == "" {
		query = `
			SELECT p.id, p.author_id, COALESCE(p.content, ''), COALESCE(p.media_url, ''), p.like_count, p.comment_count, p.created_at, p.updated_at,
			       COALESCE(pr.display_name, 'Unknown User'), COALESCE(pr.avatar_url, '')
			FROM social.posts p
			LEFT JOIN social.profiles pr ON p.author_id = pr.user_id
			WHERE p.author_id = $1
			ORDER BY p.created_at DESC
			LIMIT $2`
		rows, err = runner(ctx, r.pool).Query(ctx, query, authorID, limit)
	} else {
		query = `
			SELECT p.id, p.author_id, COALESCE(p.content, ''), COALESCE(p.media_url, ''), p.like_count, p.comment_count, p.created_at, p.updated_at,
			       COALESCE(pr.display_name, 'Unknown User'), COALESCE(pr.avatar_url, '')
			FROM social.posts p
			LEFT JOIN social.profiles pr ON p.author_id = pr.user_id
			WHERE p.author_id = $1 AND p.created_at < $3::timestamptz
			ORDER BY p.created_at DESC
			LIMIT $2`
		rows, err = runner(ctx, r.pool).Query(ctx, query, authorID, limit, cursor)
	}

	if err != nil {
		return nil, "", fmt.Errorf("listing posts by author: %w", err)
	}
	defer rows.Close()

	var posts []domain.Post
	for rows.Next() {
		var p domain.Post
		if err := rows.Scan(
			&p.ID, &p.AuthorID, &p.Content, &p.MediaURL,
			&p.LikeCount, &p.CommentCount, &p.CreatedAt, &p.UpdatedAt,
			&p.AuthorName, &p.AuthorAvatar,
		); err != nil {
			return nil, "", fmt.Errorf("scanning author post row: %w", err)
		}
		posts = append(posts, p)
	}

	var nextCursor string
	if len(posts) == limit {
		nextCursor = posts[len(posts)-1].CreatedAt.Format("2006-01-02T15:04:05.999999Z07:00")
	}

	return posts, nextCursor, nil
}

func (r *PostRepo) IncrementLikeCount(ctx context.Context, id uuid.UUID, delta int) error {
	query := `UPDATE social.posts SET like_count = like_count + $2, updated_at = NOW() WHERE id = $1`
	tag, err := runner(ctx, r.pool).Exec(ctx, query, id, delta)
	if err != nil {
		return fmt.Errorf("incrementing like count: %w", err)
	}
	// Используем tag, чтобы ошибка исчезла
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("incrementing like count: %w", domain.ErrNotFound)
	}
	return nil
}

func (r *PostRepo) IncrementCommentCount(ctx context.Context, id uuid.UUID, delta int) error {
	query := `UPDATE social.posts SET comment_count = comment_count + $2, updated_at = NOW() WHERE id = $1`

	// Выполняем запрос и получаем tag
	tag, err := runner(ctx, r.pool).Exec(ctx, query, id, delta)
	if err != nil {
		return fmt.Errorf("incrementing comment count: %w", err)
	}

	// ИСПОЛЬЗУЕМ tag: проверяем, что пост вообще существовал в базе
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("incrementing comment count: %w", domain.ErrNotFound)
	}

	return nil
}
func (r *PostRepo) LikePost(ctx context.Context, postID, userID uuid.UUID) error {
	query := `INSERT INTO social.post_likes (post_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	tag, err := runner(ctx, r.pool).Exec(ctx, query, postID, userID)
	if err != nil {
		return fmt.Errorf("liking post: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return r.IncrementLikeCount(ctx, postID, 1)
	}
	return nil
}

func (r *PostRepo) UnlikePost(ctx context.Context, postID, userID uuid.UUID) error {
	query := `DELETE FROM social.post_likes WHERE post_id = $1 AND user_id = $2`
	tag, err := runner(ctx, r.pool).Exec(ctx, query, postID, userID)
	if err != nil {
		return fmt.Errorf("unliking post: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return r.IncrementLikeCount(ctx, postID, -1)
	}
	return nil
}

func (r *PostRepo) IsLikedBy(ctx context.Context, postID, userID uuid.UUID) (bool, error) {
	query := `SELECT 1 FROM social.post_likes WHERE post_id = $1 AND user_id = $2 LIMIT 1`
	var dummy int
	err := runner(ctx, r.pool).QueryRow(ctx, query, postID, userID).Scan(&dummy)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("checking post like: %w", err)
	}
	return true, nil
}

func (r *PostRepo) CreateComment(ctx context.Context, comment *domain.PostComment) error {
	query := `
		INSERT INTO social.post_comments (post_id, author_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err := runner(ctx, r.pool).QueryRow(ctx, query,
		comment.PostID, comment.AuthorID, comment.Content,
	).Scan(&comment.ID, &comment.CreatedAt)
	if err != nil {
		return fmt.Errorf("creating comment: %w", err)
	}

	return r.IncrementCommentCount(ctx, comment.PostID, 1)
}

func (r *PostRepo) ListComments(ctx context.Context, postID uuid.UUID, cursor string, limit int) ([]domain.PostComment, string, error) {
	var query string
	var rows pgx.Rows
	var err error

	// ИСПРАВЛЕНО: Добавлен JOIN с профилями для имен в комментариях
	if cursor == "" {
		// В post_repo.go метод ListComments
		query = `
    SELECT c.id, c.post_id, c.author_id, COALESCE(c.content, ''), c.created_at,
           COALESCE(pr.display_name, 'User'), COALESCE(pr.avatar_url, '')
    FROM social.post_comments c
    LEFT JOIN social.profiles pr ON c.author_id = pr.user_id
    WHERE c.post_id = $1  -- ПРОВЕРЬТЕ ЭТУ СТРОКУ
    ORDER BY c.created_at ASC
    LIMIT $2`
		rows, err = runner(ctx, r.pool).Query(ctx, query, postID, limit)
	} else {
		query = `
			SELECT c.id, c.post_id, c.author_id, COALESCE(c.content, ''), c.created_at,
			       COALESCE(pr.display_name, 'User'), COALESCE(pr.avatar_url, '')
			FROM social.post_comments c
			LEFT JOIN social.profiles pr ON c.author_id = pr.user_id
			WHERE c.post_id = $1 AND c.created_at > $3::timestamptz
			ORDER BY c.created_at ASC
			LIMIT $2`
		rows, err = runner(ctx, r.pool).Query(ctx, query, postID, limit, cursor)
	}

	if err != nil {
		return nil, "", fmt.Errorf("listing comments: %w", err)
	}
	defer rows.Close()

	var comments []domain.PostComment
	for rows.Next() {
		var c domain.PostComment
		if err := rows.Scan(&c.ID, &c.PostID, &c.AuthorID, &c.Content, &c.CreatedAt, &c.AuthorName, &c.AuthorAvatar); err != nil {
			return nil, "", fmt.Errorf("scanning comment row: %w", err)
		}
		comments = append(comments, c)
	}

	var nextCursor string
	if len(comments) == limit {
		nextCursor = comments[len(comments)-1].CreatedAt.Format("2006-01-02T15:04:05.999999Z07:00")
	}

	return comments, nextCursor, nil
}
