// Command blog shows the types that db2go2types generates for schema.sql:
// the models package, schema.md and the tests in blog_test.go.
//
//	psql "$DATABASE_URL" -f schema.sql
//	DATABASE_URL=postgres://… go run .
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"

	"github.com/pluhsoft/db2go2types/examples/blog/models"
)

func main() {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	rows, err := conn.Query(ctx, `
		SELECT id, author_id, title, status, labels::text[], tags, "order", views, rating, meta, published_at
		FROM blog.posts WHERE status = $1`, models.PostStatusPublished)
	if err != nil {
		log.Fatal(err)
	}
	posts, err := pgx.CollectRows(rows, pgx.RowToStructByPos[models.Posts])
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range posts {
		fmt.Printf("%d %q %s %v\n", p.Id, p.Title, p.Status, p.Labels)
	}
}
