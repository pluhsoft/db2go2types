package main

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/pluhsoft/db2go2types/examples/blog/models"
	"github.com/pluhsoft/db2go2types/internal/pgtest"
)

// TestScan checks that pgx reads every column of the schema into the generated types.
func TestScan(t *testing.T) {
	db := pgtest.New(t, "schema.sql")
	ctx := context.Background()
	_, err := db.Exec(ctx, `
		INSERT INTO blog.authors (name) VALUES ('Ann');
		INSERT INTO blog.posts (author_id, title, status, labels, tags, meta, published_at)
		VALUES (1, 'Hello', 'published', '{"#A6D2FF","#BCF1A5"}', '{go,sql}', '{"lang":"en"}', now());
		INSERT INTO blog.post_likes VALUES (1, 1);`)
	if err != nil {
		t.Fatal(err)
	}

	authors, err := query[models.Authors](ctx, db, `SELECT * FROM blog.authors`)
	if err != nil || len(authors) != 1 || authors[0].Name != "Ann" || authors[0].Email != nil || authors[0].CreatedAt.IsZero() {
		t.Errorf("authors = %+v, %v", authors, err)
	}

	// Arrays of enums are selected as text[]: pgx does not know the enum array type.
	posts, err := query[models.Posts](ctx, db, `
		SELECT id, author_id, title, status, labels::text[], tags, "order", views, rating, meta, published_at
		FROM blog.posts`)
	if err != nil || len(posts) != 1 {
		t.Fatalf("posts = %+v, %v", posts, err)
	}
	p := posts[0]
	if p.Status != models.PostStatusPublished || len(p.Labels) != 2 || p.Labels[1] != models.LabelColorGreen ||
		len(p.Tags) != 2 || p.Rating != nil || p.Meta == nil || p.PublishedAt == nil {
		t.Errorf("post = %+v", p)
	}

	likes, err := query[models.PostLikes](ctx, db, `SELECT * FROM blog.post_likes`)
	if err != nil || len(likes) != 1 || likes[0].PostId != 1 {
		t.Errorf("likes = %+v, %v", likes, err)
	}
}

// query scans rows into T by column position: generated fields follow the table's column order.
func query[T any](ctx context.Context, db interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, sql string) ([]T, error) {
	rows, err := db.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[T])
}
