import re

with open("internal/adapter/postgres/post_repo.go", "r", encoding="utf-8") as f:
    content = f.read()

pattern = r"func \(r \*PostRepo\) ListFeed\(ctx context\.Context, cursor string, limit int\) \(\[\]domain\.Post, string, error\) \{.*?\treturn posts, nextCursor, nil\n\}"

newFunc = """func (r *PostRepo) ListFeed(ctx context.Context, cursor string, limit int, filter domain.PostFilter) ([]domain.Post, string, error) {
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
}"""

content = re.sub(pattern, newFunc, content, flags=re.DOTALL)

if '"fmt"' in content and '"strconv"' not in content:
    content = content.replace('"fmt"', '"fmt"\n\t"strconv"')

with open("internal/adapter/postgres/post_repo.go", "w", encoding="utf-8") as f:
    f.write(content)
