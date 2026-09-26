package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/pluhsoft/db2go2types/examples/blog/repository"
	"github.com/pluhsoft/db2go2types/internal/pgtest"
)

func ptr[T any](v T) *T { return &v }

// TestRepository runs every generated method against PostgreSQL.
func TestRepository(t *testing.T) {
	db := pgtest.New(t, "schema.sql")
	ctx := context.Background()
	authors := repository.NewAuthorsRepository(db)
	posts := repository.NewPostsRepository(db)

	// Add sends every column, so column defaults do not apply: set CreatedAt.
	author, err := authors.AddAuthors(ctx, repository.UpdateAuthorsParams{Name: "Ann", Email: ptr("ann@example.com"), CreatedAt: time.Now()}, nil)
	if err != nil {
		t.Fatal(err)
	}

	post, err := posts.AddPosts(ctx, repository.UpdatePostsParams{
		AuthorId: author.Id,
		Title:    "Hello",
		Status:   repository.PostStatusDraft,
		Labels:   []repository.LabelColor{repository.LabelColorBlue, repository.LabelColorGreen},
		Tags:     []any{"go", "sql"},
		Meta:     ptr[any](map[string]any{"lang": "en"}),
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if post.Id == 0 || post.Labels[1] != repository.LabelColorGreen || len(post.Tags) != 2 {
		t.Errorf("added post = %+v", post)
	}

	got, err := posts.GetPosts(ctx, post.Id, nil)
	if err != nil || got.Title != "Hello" || got.Status != repository.PostStatusDraft {
		t.Errorf("GetPosts = %+v, %v", got, err)
	}

	params := repository.UpdatePostsParams{AuthorId: author.Id, Title: "Hello, world", Status: repository.PostStatusPublished, Labels: got.Labels}
	updated, err := posts.UpdatePosts(ctx, params, ptr("WHERE id = 1"))
	if err != nil || updated.Status != repository.PostStatusPublished || updated.Tags != nil {
		t.Errorf("UpdatePosts = %+v, %v", updated, err)
	}

	list, err := posts.SelectPosts(ctx, ptr("WHERE status = 'published' ORDER BY id"))
	if err != nil || len(list) != 1 {
		t.Errorf("SelectPosts = %+v, %v", list, err)
	}
	if n, err := posts.CountPosts(ctx, nil); err != nil || n != 1 {
		t.Errorf("CountPosts = %d, %v", n, err)
	}
	byAuthor, err := posts.ExecuteQueryPosts(ctx,
		`SELECT id, author_id, title, status, labels::text[], tags, "order", views, rating, meta, published_at
		 FROM blog.posts WHERE author_id = $1`, author.Id)
	if err != nil || len(byAuthor) != 1 {
		t.Errorf("ExecuteQueryPosts = %+v, %v", byAuthor, err)
	}

	if err := posts.DeletePosts(ctx, ""); err == nil {
		t.Error("DeletePosts without where deleted the table")
	}
	if err := posts.DeletePosts(ctx, "WHERE id = 1"); err != nil {
		t.Fatal(err)
	}
	if _, err := posts.GetPosts(ctx, post.Id, nil); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("GetPosts after delete: %v", err)
	}
}
