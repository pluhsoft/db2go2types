// Command blog shows the code that db2go2types generates for schema.sql:
// the repository package, schema.md and the tests in blog_test.go.
//
//	psql "$DATABASE_URL" -f schema.sql
//	DATABASE_URL=postgres://… go run .
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pluhsoft/db2go2types/examples/blog/repository"
)

func main() {
	ctx := context.Background()
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	authors := repository.NewAuthorsRepository(db)
	author, err := authors.ExecuteQueryRowAuthors(ctx,
		`INSERT INTO blog.authors (name) VALUES ($1) RETURNING id, name, email, created_at`, "Ann")
	if err != nil {
		log.Fatal(err)
	}

	posts := repository.NewPostsRepository(db)
	post, err := posts.AddPosts(ctx, repository.UpdatePostsParams{
		AuthorId: author.Id,
		Title:    "Hello",
		Status:   repository.PostStatusPublished,
		Labels:   []repository.LabelColor{repository.LabelColorBlue},
	}, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("post %d by %s: %q %v\n", post.Id, author.Name, post.Title, post.Labels)
}
